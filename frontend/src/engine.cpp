#include "engine.h"
#include <fcitx/inputcontext.h>
#include <fcitx/inputpanel.h> // <--- ADD THIS IMPORTANT LINE
#include <iostream>
#include <vector>

GoVietEngine::GoVietEngine(fcitx::Instance *instance)
    : fcitx::InputMethodEngine() {
  DBusError err;
  dbus_error_init(&err);
  conn = dbus_bus_get(DBUS_BUS_SESSION, &err);
  if (dbus_error_is_set(&err)) {
    std::cerr << "DBus Connection Error: " << err.message << std::endl;
    dbus_error_free(&err);
    conn = nullptr;
  }
}

GoVietEngine::~GoVietEngine() {
  if (conn) {
    dbus_connection_unref(conn);
  }
}

std::vector<fcitx::InputMethodEntry> GoVietEngine::listInputMethods() {
  std::vector<fcitx::InputMethodEntry> entries;
  fcitx::InputMethodEntry entry("goviet", "GoViet", "vi", "goviet");
  entry.setLabel("V");
  entry.setIcon("fcitx-goviet");
  entry.setConfigurable(true);
  entries.push_back(std::move(entry));
  return entries;
}

void GoVietEngine::keyEvent(const fcitx::InputMethodEntry &entry,
                            fcitx::KeyEvent &keyEvent) {
  if (keyEvent.isRelease())
    return;

  uint32_t sym = keyEvent.key().sym();
  uint32_t state = keyEvent.key().states();
  std::string preedit, commit;

  // Call Go Backend
  bool handled = callGoBackend(sym, state, preedit, commit);

  // Get InputContext
  auto inputContext = keyEvent.inputContext();

  // Commit text first (if any)
  if (!commit.empty()) {
    // Clear preedit before committing to prevent duplicate
    inputContext->inputPanel().setClientPreedit(fcitx::Text());
    inputContext->updatePreedit();
    inputContext->commitString(commit);
  }

  // Update Preedit (only if we have preedit and haven't committed)
  if (!preedit.empty()) {
    fcitx::Text text;
    // Add underline format to indicate this is preedit, not committed text
    text.append(preedit, fcitx::TextFormatFlag::Underline);
    text.setCursor(preedit.length());
    inputContext->inputPanel().setClientPreedit(text);
    inputContext->updatePreedit();
  } else if (commit.empty()) {
    // Only clear preedit if we didn't just commit
    inputContext->inputPanel().setClientPreedit(fcitx::Text());
    inputContext->updatePreedit();
  }

  // Update UI
  inputContext->updateUserInterface(fcitx::UserInterfaceComponent::InputPanel);

  if (handled) {
    // Intercept key
    keyEvent.filterAndAccept();
  }
}

void GoVietEngine::reset(const fcitx::InputMethodEntry &,
                         fcitx::InputContextEvent &) {
  resetBackend();
}

void GoVietEngine::activate(const fcitx::InputMethodEntry &,
                            fcitx::InputContextEvent &) {
  // Reset on activate to ensure a clean state
  resetBackend();
}

void GoVietEngine::resetBackend() {
  if (!conn)
    return;

  DBusMessage *msg = dbus_message_new_method_call(
      "com.github.goviet.ime", "/Engine", "com.github.goviet.ime", "Reset");

  if (msg) {
    dbus_connection_send(conn, msg, NULL);
    dbus_message_unref(msg);
  }
}

// This function keeps the original logic
bool GoVietEngine::callGoBackend(uint32_t keysym, uint32_t modifiers,
                                 std::string &preedit, std::string &commit) {
  if (!conn)
    return false;

  DBusError err;
  dbus_error_init(&err);

  DBusMessage *msg =
      dbus_message_new_method_call("com.github.goviet.ime", "/Engine",
                                   "com.github.goviet.ime", "ProcessKey");

  if (!msg)
    return false;

  dbus_message_append_args(msg, DBUS_TYPE_UINT32, &keysym, DBUS_TYPE_UINT32,
                           &modifiers, DBUS_TYPE_INVALID);

  DBusMessage *reply =
      dbus_connection_send_with_reply_and_block(conn, msg, 200, &err);
  dbus_message_unref(msg);

  if (dbus_error_is_set(&err)) {
    dbus_error_free(&err);
    return false;
  }

  DBusMessageIter args;
  if (!dbus_message_iter_init(reply, &args)) {
    dbus_message_unref(reply);
    return false;
  }

  dbus_bool_t is_handled = false;
  char *commit_cstr = NULL;
  char *preedit_cstr = NULL;

  if (dbus_message_iter_get_arg_type(&args) == DBUS_TYPE_BOOLEAN) {
    dbus_message_iter_get_basic(&args, &is_handled);
  }

  dbus_message_iter_next(&args);
  if (dbus_message_iter_get_arg_type(&args) == DBUS_TYPE_STRING) {
    dbus_message_iter_get_basic(&args, &commit_cstr);
    if (commit_cstr)
      commit = std::string(commit_cstr);
  }

  dbus_message_iter_next(&args);
  if (dbus_message_iter_get_arg_type(&args) == DBUS_TYPE_STRING) {
    dbus_message_iter_get_basic(&args, &preedit_cstr);
    if (preedit_cstr)
      preedit = std::string(preedit_cstr);
  }

  dbus_message_unref(reply);
  return is_handled;
}