^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Changelog for package turtlesim
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

1.10.2 (2025-10-03)
-------------------
* get rid of deprecated rclcpp::spin_some() (`#183 <https://github.com/ros/ros_tutorials/issues/183>`_)
* Contributors: Alejandro Hernández Cordero

1.10.1 (2025-05-26)
-------------------
* Support Qt6 (`#170 <https://github.com/ros/ros_tutorials/issues/170>`_)
* Add icon for Kilted Kaiju (`#180 <https://github.com/ros/ros_tutorials/issues/180>`_)
* Contributors: Alejandro Hernández Cordero, Scott K Logan

1.10.0 (2025-04-25)
-------------------

1.9.2 (2024-07-09)
------------------
* Create turtlesim_msgs (`#169 <https://github.com/ros/ros_tutorials/issues/169>`_)
* Contributors: Alejandro Hernández Cordero

1.9.1 (2024-06-17)
------------------
* Add icon for Jazzy. (`#167 <https://github.com/ros/ros_tutorials/issues/167>`_)
* [teleop_turtle_key] update usage string to match keys captured by keyboard (`#165 <https://github.com/ros/ros_tutorials/issues/165>`_)
* Contributors: Marco A. Gutierrez, Mikael Arguedas

1.9.0 (2024-04-26)
------------------

1.8.2 (2024-03-28)
------------------
* Shorten the callback definition for uncrustify. (`#163 <https://github.com/ros/ros_tutorials/issues/163>`_)
* Contributors: Chris Lalancette

1.8.1 (2024-01-24)
------------------
* Use same QoS for all topic pub/subs (`#161 <https://github.com/ros/ros_tutorials/issues/161>`_)
* Contributors: Yadu

1.8.0 (2023-12-26)
------------------
* Remove all uses of ament_target_dependencies. (`#159 <https://github.com/ros/ros_tutorials/issues/159>`_)
* Contributors: Chris Lalancette

1.7.4 (2023-10-04)
------------------
* Crop galactic.png and rolling.png to 45x45. (`#158 <https://github.com/ros/ros_tutorials/issues/158>`_)
* Contributors: Jason O'Kane

1.7.3 (2023-09-07)
------------------
* Remove the unused member variable last_state\_ (`#156 <https://github.com/ros/ros_tutorials/issues/156>`_)
* Contributors: Chris Lalancette

1.7.2 (2023-06-05)
------------------
* Added common tests (`#154 <https://github.com/ros/ros_tutorials/issues/154>`_)
* Heavy cleanup of the draw_square tutorial. (`#152 <https://github.com/ros/ros_tutorials/issues/152>`_)
  * Heavy cleanup of the draw_square tutorial.
  In particular:
  1. Make it conform to the current ROS 2 style.
  2. Add in copyright information.
  3. Refactor the entire code into a class, which tidies it
  up quite a bit and removes a bunch of globals.
  4. Make sure to wait for the reset to complete before trying
  to move the turtle.
* Contributors: Alejandro Hernández Cordero, Chris Lalancette

1.7.1 (2023-05-11)
------------------
* Remove the range constraints from the holonomic parameter. (`#150 <https://github.com/ros/ros_tutorials/issues/150>`_)
* Add icon (`#148 <https://github.com/ros/ros_tutorials/issues/148>`_)
* Contributors: Chris Lalancette, Yadu

1.7.0 (2023-04-28)
------------------

1.6.0 (2023-02-14)
------------------
* Update turtlesim to C++17. (`#146 <https://github.com/ros/ros_tutorials/issues/146>`_)
* [rolling] Update maintainers - 2022-11-07 (`#145 <https://github.com/ros/ros_tutorials/issues/145>`_)
* Contributors: Audrow Nash, Chris Lalancette

1.5.0 (2022-09-13)
------------------
* Add parameter to enable holonomic motion (`#131 <https://github.com/ros/ros_tutorials/issues/131>`_)
* Add humble turtle (`#140 <https://github.com/ros/ros_tutorials/issues/140>`_)
* Contributors: Audrow Nash, Daisuke Sato

1.4.1 (2022-01-14)
------------------
* Use ``double`` when handling ``qreal orient\_`` (`#114 <https://github.com/ros/ros_tutorials/issues/114>`_)
* Add Rolling Icon (`#133 <https://github.com/ros/ros_tutorials/issues/133>`_)
* Contributors: Katherine Scott, Seulbae Kim

1.4.0 (2021-11-18)
------------------
* Update maintainers to Audrow Nash and Michael Jeronimo (`#137 <https://github.com/ros/ros_tutorials/issues/137>`_)
* Fixing deprecated subscriber callback warnings (`#134 <https://github.com/ros/ros_tutorials/issues/134>`_)
* Use rosidl_get_typesupport_target() (`#132 <https://github.com/ros/ros_tutorials/issues/132>`_)
* Print out the correct node name on startup. (`#122 <https://github.com/ros/ros_tutorials/issues/122>`_)
* Contributors: Abrar Rahman Protyasha, Audrow Nash, Chris Lalancette, Shane Loretz

1.3.3 (2021-05-21)
------------------
* Added galactic turtle icon. (`#123 <https://github.com/ros/ros_tutorials/issues/123>`_)
* Heavily cleanup teleop_turtle_key. (`#121 <https://github.com/ros/ros_tutorials/issues/121>`_)
* Contributors: Chris Lalancette, Katherine Scott

1.3.2 (2021-04-16)
------------------
* Ignore key up events in teleop_turtle_key on Windows (`#118 <https://github.com/ros/ros_tutorials/issues/118>`_)
* Contributors: Michel Hidalgo

1.3.1 (2020-12-10)
------------------
* Update maintainers (`#106 <https://github.com/ros/ros_tutorials/issues/106>`_)
* Contributors: Shane Loretz

1.3.0 (2020-09-21)
------------------
* Update goal response callback signature (`#100 <https://github.com/ros/ros_tutorials/issues/100>`_)
* Contributors: Jacob Perron

1.2.5 (2020-08-05)
------------------
* add holonomic motion for turtlesim (`#98 <https://github.com/ros/ros_tutorials/issues/98>`_)

1.2.4 (2020-06-12)
------------------
* add step value to turtlesim color parameters (`#91 <https://github.com/ros/ros_tutorials/issues/91>`_)

1.2.3 (2020-06-05)
------------------
* update Foxy turtle (`#90 <https://github.com/ros/ros_tutorials/issues/90>`_)

1.2.2 (2020-06-03)
------------------
* add Foxy turtle (`#89 <https://github.com/ros/ros_tutorials/issues/89>`_)

1.2.1 (2020-06-02)
------------------
* Fix Qt deprecation warning (`#88 <https://github.com/ros/ros_tutorials/issues/88>`_)

1.2.0 (2020-05-01)
------------------
* Replace deprecated launch_ros usage (`#84 <https://github.com/ros/ros_tutorials/issues/84>`_)

1.1.1 (2020-04-16)
------------------
* catch reference to fix -Wcatch-value warning (`#78 <https://github.com/ros/ros_tutorials/issues/78>`_)

1.1.0 (2019-11-12)
------------------
* Eloquent Elusor turtle icon (`#77 <https://github.com/ros/ros_tutorials/issues/77>`_)

1.0.2 (2019-10-23)
------------------
* separate background color from drawn paths, trigger repaint on parameter changes (`#75 <https://github.com/ros/ros_tutorials/issues/75>`_)
* add descriptor information for background color parameters (`#73 <https://github.com/ros/ros_tutorials/issues/73>`_)
* Fix Windows compiler warning (`#69 <https://github.com/ros/ros_tutorials/issues/69>`_)
* Change log messages to use 'goal' instead of 'action' (`#67 <https://github.com/ros/ros_tutorials/issues/67>`_)

1.0.1 (2019-10-02)
------------------
* fix mimic tutorial node (`#65 <https://github.com/ros/ros_tutorials/issues/65>`_)
* fix syntax error in teleop_turtle_key.cpp on Windows (`#66 <https://github.com/ros/ros_tutorials/issues/66>`_)
* add RotateAbsolute action (`#62 <https://github.com/ros/ros_tutorials/issues/62>`_)
* fix typo in error message (`#64 <https://github.com/ros/ros_tutorials/issues/64>`_)

1.0.0 (2019-09-24)
------------------
* replace images with ROS 2 turtles (`#60 <https://github.com/ros/ros_tutorials/issues/60>`_)
* add shortcut to quit teleop (`#58 <https://github.com/ros/ros_tutorials/issues/58>`_)
* fix compiler warnings on Windows (`#57 <https://github.com/ros/ros_tutorials/issues/57>`_)
* add support for Windows (`#56 <https://github.com/ros/ros_tutorials/issues/56>`_)
* various fixes for ROS 2 (`#55 <https://github.com/ros/ros_tutorials/issues/55>`_)
* turtlesim for ROS 2 (`#53 <https://github.com/ros/ros_tutorials/issues/53>`_)

0.9.1 (2019-03-04)
------------------
* change formula to avoid rounding with extreme input values (`#51 <https://github.com/ros/ros_tutorials/issues/51>`_)
* keep theta in the desired interval (`#46 <https://github.com/ros/ros_tutorials/issues/46>`_)

0.9.0 (2018-04-11)
------------------
* add melodic turtle (`#41 <https://github.com/ros/ros_tutorials/issues/41>`_)

0.8.1 (2017-07-27)
------------------
* theta ranges from -pi to +pi (`#31 <https://github.com/ros/ros_tutorials/issues/31>`_)

0.8.0 (2017-03-10)
------------------
* add lunar turtle (`#39 <https://github.com/ros/ros_tutorials/pull/39>`_)

0.7.1 (2016-10-24)
------------------
* check pen_on\_ when processing teleport requests (`#35 <https://github.com/ros/ros_tutorials/pull/35>`_)

0.7.0 (2016-03-18)
------------------
* add kinetic image
* update to Qt5
* fix size of Jade image to not exceed other images in order to not get positioned incorrectly
* fix compiler warnings

0.6.1 (2015-09-19)
------------------
* update the coordinate system in /spawn service for consistency (`#25 <https://github.com/ros/ros_tutorials/pull/25>`_)

0.6.0 (2015-05-21)
------------------
* add jade turtle (`#22 <https://github.com/ros/ros_tutorials/pull/22>`_)

0.5.3 (2015-05-04)
------------------

0.5.2 (2014-12-23)
------------------

0.5.1 (2014-05-08)
------------------

0.5.0 (2014-05-07)
------------------
* add indigo turtle
* add disabled code to easily spawn all available turtle types

0.4.3 (2014-01-07)
------------------

0.4.2 (2013-10-04)
------------------
* fix missing install of hydro.svg (`#12 <https://github.com/ros/ros_tutorials/issues/12>`_)

0.4.1 (2013-09-11)
------------------
* add hydro image to turtlesim

0.4.0 (2013-09-06)
------------------
* Adding png version of hydro for wiki linking
* TurtleApp accepts argc by reference
* Restoring all the changes appropriate for Hydro

0.3.13 (2013-08-21)
-------------------
* TurtleApp accepts argc by reference
* add hydro image to turtlesim
* remove mainpage.dox

0.3.12 (2013-03-29)
-------------------
* reverting velocity -> twist for groovy
* Revert "chaning command_velocity to cmd_vel" for groovy
  This reverts commit 96e5174d3a5c961b6e1195b90b4024e2858df010.
* Revert "adding geometry_msgs dependency in package.xml and CMakelist" for groovy
  This reverts commit c7ac1155d70269909b55af03d13fe2e089d6215d.
* Revert "alaphabetic order" for groovy
  This reverts commit f928765ed08773517c195b74c55231c0e4fcc5e5.

0.3.11 (2013-03-21)
-------------------
* update email in package.xml

0.3.10 (2013-03-08)
-------------------
* Fix a moc generation error with boost >= 1.48
  See:
  https://bugreports.qt-project.org/browse/QTBUG-22829
* Revert "Merge pull request `#6 <https://github.com/ros/ros_tutorials/issues/6>`_ from ros/fix_qt_moc"
  This reverts commit 0e11b41ac53aad0e043b77d4d5950889245eaceb, reversing
  changes made to fc19df449d9ac297e8ab829ff22e99323c33ae93.
* Revert "fix missing include (regression of `#5 <https://github.com/ros/ros_tutorials/issues/5>`_)"
  This reverts commit 546dabe05c00e87296952cb2ca655e01895bd5ed.
* fix missing include (regression of `#5 <https://github.com/ros/ros_tutorials/issues/5>`_)
* Fix a mod generation error with boost >= 1.48
  See:
  https://bugreports.qt-project.org/browse/QTBUG-22829
* alaphabetic order
* adding geometry_msgs dependency in package.xml and CMakelist
* chaning command_velocity to cmd_vel
* remove turtlesim velocity and use Twist msg

0.3.9 (2012-12-21)
------------------
* add groovy turtle
* modified dep type of catkin

0.3.8 (2012-12-13)
------------------
* add missing downstream depend
* switched from langs to message_* packages

0.3.7 (2012-12-06)
------------------

0.3.6 (2012-10-30)
------------------
* fix catkin function order

0.3.5 (2012-10-18)
------------------

0.3.4 (2012-10-06)
------------------

0.3.3 (2012-10-05)
------------------
* fixed missing genmsg stuff
* updated to latest catkin
* added package.xml files

0.3.2 (2012-09-05)
------------------
* updated catkin variables
* updated pkg-config in manifest.xml

0.3.1 (2012-09-03)
------------------
* use install destination variables, removed manual installation of manifests

0.3.0 (2012-08-29)
------------------
* updated to current catkin

0.2.20 (2013-02-08)
-------------------
* fixed compilation on platforms with different qreal type

0.2.19 (2012-06-15 03:13:40 +0000)
----------------------------------
* make find_package REQUIRED
* removed obsolete catkin tag from manifest files
* added missing install of turtlesim images
* using fuerte image in turtlesim
* fuerte icon
* remove old Makefiles and bump to 0.2.13
* fix find boost component for turtlesim
* change deps for turtlesim from wx to qt
* migrate turtlesim from wx to qt
* updated export for messages/catkin
* add missing libs for oneiric
* add missing dependency on wx, and take out conditional build logic from turtlesim
* conditionally build based on wx, for now
* turn on turtlesim
* adding <catkin/>, removing depends and platform tags
* remove old rosbuild2 stuff
* adios rosbuild2 in manifests
* changed number of turtles to a #define to prevent future mistakes with adding new turtles
* electric turtle
* rosbuild2/windows tweaks, they keep on comin'
* rosbuild2 taking shape.
* rosbuild2 taking shape
* moving teleop keyboard into turtlesim to remove tutorial deps on keyboard
* diamondback
* Added Ubuntu platform tags
* fix to actually paint on OSX
* Only update the path image every 3 frames, because ConvertToImage on a 500x500 bitmap is somehow very expensive
* Move bitmap->image conversion outside of loop (that was boneheaded)
* Add color sensor to turtles
* Switch turtlesim to x-forward (theta=0 now faces to the right)
* Optionally name your turtles yourself
* Fix coordinate system
* adding a little more description to manifest
* Add absolute and relative teleport service calls
* changing turtlesim to turtlesim_node for tutorial clarity
* * Multi-turtle support
  * turtle_pose and command_velocity now exist per-turtle.  turtle_pose has been renamed "pose"
  * "spawn" service call to spawn a new turtle, which returns the turtle name
  * "kill" service call, to kill a turtle by name
  * Switch to "meters" as the distance unit, where 1 meter is defined as the height of the turtle
* adding export to manifest
* Change default background/pen colors
* Randomly choose one of the 3 turtles
* 3 turtle set by metamanda
* throttling refresh rate so that xorg doesn't use all the cpu
* adding debug statements
* the drawing file used to create turtle.png
* new turtle made by melonee
* Apply Melonee's diff to set the background color parameters on the param server at startup
* Add error output if the turtle hits the wall
* Add turtlesim to the ros_tutorials stack
