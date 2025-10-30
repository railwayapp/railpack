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
#include <signal.h>
#include <stdio.h>

#include <functional>
#include <stdexcept>
#include <thread>

#include <rclcpp/rclcpp.hpp>
#include <rclcpp_action/rclcpp_action.hpp>
#include <geometry_msgs/msg/twist.hpp>
#include <turtlesim_msgs/action/rotate_absolute.hpp>

#ifdef _WIN32
# include <windows.h>  // NO LINT
#else
# include <termios.h>  // NO LINT
# include <unistd.h>  // NO LINT
#endif

#include "turtlesim/qos.hpp"

static constexpr char KEYCODE_RIGHT = 0x43;
static constexpr char KEYCODE_LEFT = 0x44;
static constexpr char KEYCODE_UP = 0x41;
static constexpr char KEYCODE_DOWN = 0x42;
static constexpr char KEYCODE_B = 0x62;
static constexpr char KEYCODE_C = 0x63;
static constexpr char KEYCODE_D = 0x64;
static constexpr char KEYCODE_E = 0x65;
static constexpr char KEYCODE_F = 0x66;
static constexpr char KEYCODE_G = 0x67;
static constexpr char KEYCODE_Q = 0x71;
static constexpr char KEYCODE_R = 0x72;
static constexpr char KEYCODE_T = 0x74;
static constexpr char KEYCODE_V = 0x76;

bool running = true;

class KeyboardReader final
{
public:
  KeyboardReader()
  {
#ifdef _WIN32
    hstdin_ = GetStdHandle(STD_INPUT_HANDLE);
    if (hstdin_ == INVALID_HANDLE_VALUE) {
      throw std::runtime_error("Failed to get stdin handle");
    }
    if (!GetConsoleMode(hstdin_, &old_mode_)) {
      throw std::runtime_error("Failed to get old console mode");
    }
    DWORD new_mode = ENABLE_PROCESSED_INPUT;  // for Ctrl-C processing
    if (!SetConsoleMode(hstdin_, new_mode)) {
      throw std::runtime_error("Failed to set new console mode");
    }
#else
    // get the console in raw mode
    if (tcgetattr(0, &cooked_) < 0) {
      throw std::runtime_error("Failed to get old console mode");
    }
    struct termios raw;
    memcpy(&raw, &cooked_, sizeof(struct termios));
    raw.c_lflag &= ~(ICANON | ECHO);
    // Setting a new line, then end of file
    raw.c_cc[VEOL] = 1;
    raw.c_cc[VEOF] = 2;
    raw.c_cc[VTIME] = 1;
    raw.c_cc[VMIN] = 0;
    if (tcsetattr(0, TCSANOW, &raw) < 0) {
      throw std::runtime_error("Failed to set new console mode");
    }
#endif
  }

  char readOne()
  {
    char c = 0;

#ifdef _WIN32
    INPUT_RECORD record;
    DWORD num_read;
    switch (WaitForSingleObject(hstdin_, 100)) {
      case WAIT_OBJECT_0:
        if (!ReadConsoleInput(hstdin_, &record, 1, &num_read)) {
          throw std::runtime_error("Read failed");
        }

        if (record.EventType != KEY_EVENT || !record.Event.KeyEvent.bKeyDown) {
          break;
        }

        if (record.Event.KeyEvent.wVirtualKeyCode == VK_LEFT) {
          c = KEYCODE_LEFT;
        } else if (record.Event.KeyEvent.wVirtualKeyCode == VK_UP) {
          c = KEYCODE_UP;
        } else if (record.Event.KeyEvent.wVirtualKeyCode == VK_RIGHT) {
          c = KEYCODE_RIGHT;
        } else if (record.Event.KeyEvent.wVirtualKeyCode == VK_DOWN) {
          c = KEYCODE_DOWN;
        } else if (record.Event.KeyEvent.wVirtualKeyCode == 0x42) {
          c = KEYCODE_B;
        } else if (record.Event.KeyEvent.wVirtualKeyCode == 0x43) {
          c = KEYCODE_C;
        } else if (record.Event.KeyEvent.wVirtualKeyCode == 0x44) {
          c = KEYCODE_D;
        } else if (record.Event.KeyEvent.wVirtualKeyCode == 0x45) {
          c = KEYCODE_E;
        } else if (record.Event.KeyEvent.wVirtualKeyCode == 0x46) {
          c = KEYCODE_F;
        } else if (record.Event.KeyEvent.wVirtualKeyCode == 0x47) {
          c = KEYCODE_G;
        } else if (record.Event.KeyEvent.wVirtualKeyCode == 0x51) {
          c = KEYCODE_Q;
        } else if (record.Event.KeyEvent.wVirtualKeyCode == 0x52) {
          c = KEYCODE_R;
        } else if (record.Event.KeyEvent.wVirtualKeyCode == 0x54) {
          c = KEYCODE_T;
        } else if (record.Event.KeyEvent.wVirtualKeyCode == 0x56) {
          c = KEYCODE_V;
        }
        break;

      case WAIT_TIMEOUT:
        break;
    }

#else
    int rc = read(0, &c, 1);
    if (rc < 0) {
      throw std::runtime_error("read failed");
    }
#endif

    return c;
  }

  ~KeyboardReader()
  {
#ifdef _WIN32
    SetConsoleMode(hstdin_, old_mode_);
#else
    tcsetattr(0, TCSANOW, &cooked_);
#endif
  }

private:
#ifdef _WIN32
  HANDLE hstdin_;
  DWORD old_mode_;
#else
  struct termios cooked_;
#endif
};

class TeleopTurtle final
{
public:
  TeleopTurtle()
  {
    nh_ = rclcpp::Node::make_shared("teleop_turtle");
    nh_->declare_parameter("scale_angular", 2.0);
    nh_->declare_parameter("scale_linear", 2.0);

    twist_pub_ = nh_->create_publisher<geometry_msgs::msg::Twist>(
      "turtle1/cmd_vel",
      turtlesim::topic_qos());
    rotate_absolute_client_ = rclcpp_action::create_client<turtlesim_msgs::action::RotateAbsolute>(
      nh_,
      "turtle1/rotate_absolute");
  }

  int keyLoop()
  {
    char c;

    std::thread{std::bind(&TeleopTurtle::spin, this)}.detach();

    puts("Reading from keyboard");
    puts("---------------------------");
    puts("Use arrow keys to move the turtle.");
    puts("Use g|b|v|c|d|e|r|t keys to rotate to absolute orientations. 'f' to cancel a rotation.");
    puts("'q' to quit.");

    while (running) {
      // get the next event from the keyboard
      try {
        c = input_.readOne();
      } catch (const std::runtime_error &) {
        perror("read():");
        return -1;
      }

      double linear = 0.0;
      double angular = 0.0;

      RCLCPP_DEBUG(nh_->get_logger(), "value: 0x%02X\n", c);

      switch (c) {
        case KEYCODE_LEFT:
          RCLCPP_DEBUG(nh_->get_logger(), "LEFT");
          angular = 1.0;
          break;
        case KEYCODE_RIGHT:
          RCLCPP_DEBUG(nh_->get_logger(), "RIGHT");
          angular = -1.0;
          break;
        case KEYCODE_UP:
          RCLCPP_DEBUG(nh_->get_logger(), "UP");
          linear = 1.0;
          break;
        case KEYCODE_DOWN:
          RCLCPP_DEBUG(nh_->get_logger(), "DOWN");
          linear = -1.0;
          break;
        case KEYCODE_G:
          RCLCPP_DEBUG(nh_->get_logger(), "G");
          sendGoal(0.0f);
          break;
        case KEYCODE_T:
          RCLCPP_DEBUG(nh_->get_logger(), "T");
          sendGoal(0.7854f);
          break;
        case KEYCODE_R:
          RCLCPP_DEBUG(nh_->get_logger(), "R");
          sendGoal(1.5708f);
          break;
        case KEYCODE_E:
          RCLCPP_DEBUG(nh_->get_logger(), "E");
          sendGoal(2.3562f);
          break;
        case KEYCODE_D:
          RCLCPP_DEBUG(nh_->get_logger(), "D");
          sendGoal(3.1416f);
          break;
        case KEYCODE_C:
          RCLCPP_DEBUG(nh_->get_logger(), "C");
          sendGoal(-2.3562f);
          break;
        case KEYCODE_V:
          RCLCPP_DEBUG(nh_->get_logger(), "V");
          sendGoal(-1.5708f);
          break;
        case KEYCODE_B:
          RCLCPP_DEBUG(nh_->get_logger(), "B");
          sendGoal(-0.7854f);
          break;
        case KEYCODE_F:
          RCLCPP_DEBUG(nh_->get_logger(), "F");
          cancelGoal();
          break;
        case KEYCODE_Q:
          RCLCPP_DEBUG(nh_->get_logger(), "quit");
          running = false;
          break;
        default:
          // This can happen if the read returned when there was no data, or
          // another key was pressed.  In these cases, just silently ignore the
          // press.
          break;
      }

      if (running && (linear != 0.0 || angular != 0.0)) {
        geometry_msgs::msg::Twist twist;
        twist.angular.z = nh_->get_parameter("scale_angular").as_double() * angular;
        twist.linear.x = nh_->get_parameter("scale_linear").as_double() * linear;
        twist_pub_->publish(twist);
      }
    }

    return 0;
  }

private:
  void spin()
  {
    rclcpp::spin(nh_);
  }

  void sendGoal(float theta)
  {
    using Rotate = turtlesim_msgs::action::RotateAbsolute;
    auto goal = Rotate::Goal();
    goal.theta = theta;
    auto send_goal_options = rclcpp_action::Client<Rotate>::SendGoalOptions();
    send_goal_options.goal_response_callback =
      [this](rclcpp_action::ClientGoalHandle<Rotate>::SharedPtr goal_handle)
      {
        RCLCPP_DEBUG(nh_->get_logger(), "Goal response received");
        this->goal_handle_ = goal_handle;
      };
    rotate_absolute_client_->async_send_goal(goal, send_goal_options);
  }

  void cancelGoal()
  {
    if (goal_handle_) {
      RCLCPP_DEBUG(nh_->get_logger(), "Sending cancel request");
      try {
        rotate_absolute_client_->async_cancel_goal(goal_handle_);
      } catch (...) {
        // This can happen if the goal has already terminated and expired
      }
    }
  }

  rclcpp::Node::SharedPtr nh_;
  rclcpp::Publisher<geometry_msgs::msg::Twist>::SharedPtr twist_pub_;
  rclcpp_action::Client<turtlesim_msgs::action::RotateAbsolute>::SharedPtr rotate_absolute_client_;
  rclcpp_action::ClientGoalHandle<turtlesim_msgs::action::RotateAbsolute>::SharedPtr goal_handle_;

  KeyboardReader input_;
};

#ifdef _WIN32
BOOL WINAPI quit(DWORD ctrl_type)
{
  (void)ctrl_type;
  running = false;
  return true;
}
#else
void quit(int sig)
{
  (void)sig;
  running = false;
}
#endif

int main(int argc, char ** argv)
{
  rclcpp::init(argc, argv);

#ifdef _WIN32
  SetConsoleCtrlHandler(quit, TRUE);
#else
  signal(SIGINT, quit);
#endif

  TeleopTurtle teleop_turtle;

  int rc = teleop_turtle.keyLoop();

  rclcpp::shutdown();

  return rc;
}
