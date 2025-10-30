// Copyright (c) 2009, Willow Garage, Inc.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
//    * Redistributions of source code must retain the above copyright
//      notice, this list of conditions and the following disclaimer.
//
//    * Redistributions in binary form must reproduce the above copyright
//      notice, this list of conditions and the following disclaimer in the
//      documentation and/or other materials provided with the distribution.
//
//    * Neither the name of the Willow Garage nor the names of its
//      contributors may be used to endorse or promote products derived from
//      this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

#include <chrono>
#include <cmath>
#include <functional>
#include <memory>

#include <geometry_msgs/msg/twist.hpp>
#include <rclcpp/rclcpp.hpp>
#include <std_srvs/srv/empty.hpp>
#include <turtlesim_msgs/msg/pose.hpp>

#include "turtlesim/qos.hpp"

#define PI 3.141592f

class DrawSquare final : public rclcpp::Node
{
public:
  explicit DrawSquare(const rclcpp::NodeOptions & options = rclcpp::NodeOptions())
  : rclcpp::Node("draw_square", options)
  {
    const rclcpp::QoS qos = turtlesim::topic_qos();
    twist_pub_ = this->create_publisher<geometry_msgs::msg::Twist>("turtle1/cmd_vel", qos);

    pose_sub_ =
      this->create_subscription<turtlesim_msgs::msg::Pose>(
      "turtle1/pose", qos, std::bind(&DrawSquare::poseCallback, this, std::placeholders::_1));

    reset_client_ = this->create_client<std_srvs::srv::Empty>("reset");

    timer_ = this->create_wall_timer(std::chrono::milliseconds(16), [this]() {timerCallback();});

    auto empty = std::make_shared<std_srvs::srv::Empty::Request>();
    reset_result_ = reset_client_->async_send_request(empty).future;
  }

private:
  enum State
  {
    FORWARD,
    STOP_FORWARD,
    TURN,
    STOP_TURN,
  };

  void poseCallback(const turtlesim_msgs::msg::Pose & pose)
  {
    current_pose_ = pose;
    first_pose_set_ = true;
  }

  bool hasReachedGoal()
  {
    return fabsf(current_pose_.x - goal_pose_.x) < 0.1 &&
           fabsf(current_pose_.y - goal_pose_.y) < 0.1 &&
           fabsf(current_pose_.theta - goal_pose_.theta) < 0.01;
  }

  bool hasStopped()
  {
    return current_pose_.angular_velocity < 0.0001 && current_pose_.linear_velocity < 0.0001;
  }

  void printGoal()
  {
    RCLCPP_INFO(
      this->get_logger(), "New goal [%f %f, %f]", goal_pose_.x, goal_pose_.y, goal_pose_.theta);
  }

  void commandTurtle(float linear, float angular)
  {
    geometry_msgs::msg::Twist twist;
    twist.linear.x = linear;
    twist.angular.z = angular;
    twist_pub_->publish(twist);
  }

  void stopForward()
  {
    if (hasStopped()) {
      RCLCPP_INFO(this->get_logger(), "Reached goal");
      state_ = TURN;
      goal_pose_.x = current_pose_.x;
      goal_pose_.y = current_pose_.y;
      goal_pose_.theta = fmod(current_pose_.theta + PI / 2.0f, 2.0f * PI);
      // wrap goal_pose_.theta to [-pi, pi)
      if (goal_pose_.theta >= PI) {
        goal_pose_.theta -= 2.0f * PI;
      }
      printGoal();
    } else {
      commandTurtle(0, 0);
    }
  }

  void stopTurn()
  {
    if (hasStopped()) {
      RCLCPP_INFO(this->get_logger(), "Reached goal");
      state_ = FORWARD;
      goal_pose_.x = cos(current_pose_.theta) * 2 + current_pose_.x;
      goal_pose_.y = sin(current_pose_.theta) * 2 + current_pose_.y;
      goal_pose_.theta = current_pose_.theta;
      printGoal();
    } else {
      commandTurtle(0, 0);
    }
  }

  void forward()
  {
    if (hasReachedGoal()) {
      state_ = STOP_FORWARD;
      commandTurtle(0, 0);
    } else {
      commandTurtle(1.0f, 0);
    }
  }

  void turn()
  {
    if (hasReachedGoal()) {
      state_ = STOP_TURN;
      commandTurtle(0, 0);
    } else {
      commandTurtle(0, 0.4f);
    }
  }

  void timerCallback()
  {
    if (!reset_result_.valid()) {
      return;
    }

    if (!first_pose_set_) {
      return;
    }

    if (!first_goal_set_) {
      first_goal_set_ = true;
      state_ = FORWARD;
      goal_pose_.x = cos(current_pose_.theta) * 2 + current_pose_.x;
      goal_pose_.y = sin(current_pose_.theta) * 2 + current_pose_.y;
      goal_pose_.theta = current_pose_.theta;
      printGoal();
    }

    if (state_ == FORWARD) {
      forward();
    } else if (state_ == STOP_FORWARD) {
      stopForward();
    } else if (state_ == TURN) {
      turn();
    } else if (state_ == STOP_TURN) {
      stopTurn();
    }
  }

  turtlesim_msgs::msg::Pose current_pose_;
  turtlesim_msgs::msg::Pose goal_pose_;
  bool first_goal_set_ = false;
  bool first_pose_set_ = false;
  State state_ = FORWARD;

  rclcpp::Publisher<geometry_msgs::msg::Twist>::SharedPtr twist_pub_;
  rclcpp::Subscription<turtlesim_msgs::msg::Pose>::SharedPtr pose_sub_;
  rclcpp::Client<std_srvs::srv::Empty>::SharedPtr reset_client_;
  rclcpp::Client<std_srvs::srv::Empty>::SharedFuture reset_result_;
  rclcpp::TimerBase::SharedPtr timer_;
};

int main(int argc, char ** argv)
{
  rclcpp::init(argc, argv);

  auto nh = std::make_shared<DrawSquare>();

  rclcpp::spin(nh);

  rclcpp::shutdown();

  return 0;
}
