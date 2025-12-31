#ifndef GOVIET_ENGINE_H
#define GOVIET_ENGINE_H

#include <vector>
#include <fcitx/inputmethodengine.h>
#include <fcitx/addonfactory.h>
#include <fcitx/instance.h>
#include <dbus/dbus.h>

class GoVietEngine : public fcitx::InputMethodEngine {
public:
    GoVietEngine(fcitx::Instance *instance);
    ~GoVietEngine();

    std::vector<fcitx::InputMethodEntry> listInputMethods() override;
    void keyEvent(const fcitx::InputMethodEntry& entry, fcitx::KeyEvent& keyEvent) override;

private:
    DBusConnection* conn;
    bool callGoBackend(uint32_t keysym, uint32_t modifiers, std::string &preedit, std::string &commit);
};

class GoVietEngineFactory : public fcitx::AddonFactory {
public:
    fcitx::AddonInstance *create(fcitx::AddonManager *manager) override;
};

#endif