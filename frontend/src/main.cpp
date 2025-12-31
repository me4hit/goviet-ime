#include "engine.h"
#include <fcitx/addonmanager.h>

// Register Factory with Fcitx5
FCITX_ADDON_FACTORY(GoVietEngineFactory);

fcitx::AddonInstance *GoVietEngineFactory::create(fcitx::AddonManager *manager) {
    return new GoVietEngine(manager->instance());
}