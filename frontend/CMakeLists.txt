cmake_minimum_required(VERSION 3.10)
project(goviet-ime)

# --- UPDATE: Set to C++20 ---
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(ECM REQUIRED)
set(CMAKE_MODULE_PATH ${ECM_MODULE_PATH})
include(GNUInstallDirs)

find_package(PkgConfig REQUIRED)
# Add Fcitx5Config component to ensure config headers are loaded
pkg_check_modules(FCITX5 REQUIRED Fcitx5Core Fcitx5Module Fcitx5Config)
pkg_check_modules(DBUS REQUIRED dbus-1)

set(GOVIET_LIB_NAME "libgo-viet-engine")
configure_file(goviet.conf.in goviet.conf)

add_library(go-viet-engine MODULE
    src/main.cpp
    src/engine.cpp
)

target_include_directories(go-viet-engine PRIVATE
    ${FCITX5_INCLUDE_DIRS}
    ${DBUS_INCLUDE_DIRS}
)

target_link_libraries(go-viet-engine
    ${FCITX5_LIBRARIES}
    ${DBUS_LIBRARIES}
)

target_compile_options(go-viet-engine PRIVATE 
    ${FCITX5_CFLAGS_OTHER} 
    ${DBUS_CFLAGS_OTHER}
)

# --- INSTALLATION SECTION ---

# Directly target Arch system directories
install(TARGETS go-viet-engine DESTINATION "/usr/lib/fcitx5")
install(FILES "${CMAKE_CURRENT_BINARY_DIR}/goviet.conf" DESTINATION "/usr/share/fcitx5/addon")
install(FILES "goviet-im.conf" DESTINATION "/usr/share/fcitx5/inputmethod" RENAME "goviet.conf")