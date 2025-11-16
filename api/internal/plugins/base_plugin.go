// Package plugins provides the plugin system for StreamSpace API.
//
// The base_plugin component provides default no-op implementations of the
// PluginHandler interface, following the "convention over configuration" pattern.
//
// Design Pattern - Embedding for Selective Override:
//
// Instead of requiring plugins to implement all 13 lifecycle hook methods,
// plugins can embed BasePlugin and only override the hooks they need:
//
//	type SlackPlugin struct {
//	    plugins.BasePlugin  // Embeds all default implementations
//	}
//
//	// Only override hooks you need
//	func (p *SlackPlugin) OnLoad(ctx *PluginContext) error {
//	    // Custom initialization
//	    return nil
//	}
//
//	func (p *SlackPlugin) OnSessionCreated(ctx *PluginContext, session interface{}) error {
//	    // Send Slack notification
//	    return nil
//	}
//
//	// All other hooks (OnUserLogin, OnSessionDeleted, etc.) use default no-op
//
// This pattern:
//   - Reduces boilerplate code in plugins
//   - Makes plugins easier to write and maintain
//   - Provides forward compatibility when new hooks are added
//   - Follows Go's composition-over-inheritance model
//
// Hook Categories:
//
//  1. Plugin Lifecycle:
//     - OnLoad: Plugin initialization
//     - OnUnload: Plugin cleanup
//     - OnEnable: Plugin enabled
//     - OnDisable: Plugin disabled
//
//  2. Session Hooks:
//     - OnSessionCreated, OnSessionStarted, OnSessionStopped
//     - OnSessionHibernated, OnSessionWoken, OnSessionDeleted
//
//  3. User Hooks:
//     - OnUserCreated, OnUserUpdated, OnUserDeleted
//     - OnUserLogin, OnUserLogout
//
// Built-in Plugin Registry:
//
// This file also provides a global registry for built-in plugins. Built-in
// plugins are compiled into the binary and automatically discovered at startup:
//
//	// In plugin code (e.g., slack_plugin.go)
//	func init() {
//	    plugins.RegisterBuiltinPlugin("slack", &SlackPlugin{})
//	}
//
//	// Runtime automatically loads all registered built-ins
//
// Built-in vs Dynamic:
//   - Built-in: Compiled into binary, always available, faster startup
//   - Dynamic: Loaded from .so files, can be added without recompile
package plugins

import "fmt"

// BasePlugin provides default no-op implementations for the PluginHandler interface.
//
// Plugins can embed this struct to inherit default implementations and only
// override the lifecycle hooks they actually need.
//
// Benefits:
//   - Reduces boilerplate: Don't implement unused hooks
//   - Forward compatibility: New hooks added to interface don't break existing plugins
//   - Convention over configuration: Most plugins only need 2-3 hooks
//
// Usage:
//
//	type MyPlugin struct {
//	    plugins.BasePlugin
//	}
//
//	// Override only what you need
//	func (p *MyPlugin) OnLoad(ctx *PluginContext) error {
//	    // Initialize plugin
//	    return nil
//	}
//
// All hook methods return nil (success) by default.
type BasePlugin struct {
	// Name is the plugin identifier.
	// Set during registration, not by plugin code.
	Name string
}

// Plugin Lifecycle Hooks - Default no-op implementations

// OnLoad is called when the plugin is first loaded.
// Default: no-op. Override to initialize plugin resources.
func (p *BasePlugin) OnLoad(ctx *PluginContext) error {
	return nil
}

// OnUnload is called when the plugin is being unloaded.
// Default: no-op. Override to clean up plugin resources.
func (p *BasePlugin) OnUnload(ctx *PluginContext) error {
	return nil
}

// OnEnable is called when the plugin is enabled.
// Default: no-op. Override to start plugin services.
func (p *BasePlugin) OnEnable(ctx *PluginContext) error {
	return nil
}

// OnDisable is called when the plugin is disabled.
// Default: no-op. Override to pause plugin services.
func (p *BasePlugin) OnDisable(ctx *PluginContext) error {
	return nil
}

// Session Event Hooks - Default no-op implementations

// OnSessionCreated is called when a new session is created.
// Default: no-op. Override to track session creation or send notifications.
func (p *BasePlugin) OnSessionCreated(ctx *PluginContext, session interface{}) error {
	return nil
}

// OnSessionStarted is called when a session starts (transitions to running).
// Default: no-op. Override to react to session startup.
func (p *BasePlugin) OnSessionStarted(ctx *PluginContext, session interface{}) error {
	return nil
}

// OnSessionStopped is called when a session stops.
// Default: no-op. Override to clean up session-specific resources.
func (p *BasePlugin) OnSessionStopped(ctx *PluginContext, session interface{}) error {
	return nil
}

// OnSessionHibernated is called when a session is hibernated (scale to zero).
// Default: no-op. Override to react to hibernation.
func (p *BasePlugin) OnSessionHibernated(ctx *PluginContext, session interface{}) error {
	return nil
}

// OnSessionWoken is called when a hibernated session wakes up.
// Default: no-op. Override to react to session wake.
func (p *BasePlugin) OnSessionWoken(ctx *PluginContext, session interface{}) error {
	return nil
}

// OnSessionDeleted is called when a session is permanently deleted.
// Default: no-op. Override to clean up or log deletion.
func (p *BasePlugin) OnSessionDeleted(ctx *PluginContext, session interface{}) error {
	return nil
}

// User Event Hooks - Default no-op implementations

// OnUserCreated is called when a new user account is created.
// Default: no-op. Override to provision user-specific resources.
func (p *BasePlugin) OnUserCreated(ctx *PluginContext, user interface{}) error {
	return nil
}

// OnUserUpdated is called when a user profile is updated.
// Default: no-op. Override to sync user data.
func (p *BasePlugin) OnUserUpdated(ctx *PluginContext, user interface{}) error {
	return nil
}

// OnUserDeleted is called when a user account is deleted.
// Default: no-op. Override to clean up user data.
func (p *BasePlugin) OnUserDeleted(ctx *PluginContext, user interface{}) error {
	return nil
}

// OnUserLogin is called when a user logs in.
// Default: no-op. Override to track login events.
func (p *BasePlugin) OnUserLogin(ctx *PluginContext, user interface{}) error {
	return nil
}

// OnUserLogout is called when a user logs out.
// Default: no-op. Override to clean up session data.
func (p *BasePlugin) OnUserLogout(ctx *PluginContext, user interface{}) error {
	return nil
}

// Built-in Plugin Registry

// builtinPlugins stores plugins compiled into the binary.
//
// Built-in plugins are registered via init() functions and automatically
// discovered by the plugin runtime at startup.
var builtinPlugins = make(map[string]PluginHandler)

// RegisterBuiltinPlugin registers a plugin as built-in.
//
// This should be called from init() functions in plugin packages:
//
//	func init() {
//	    plugins.RegisterBuiltinPlugin("slack", &SlackPlugin{})
//	}
//
// Thread Safety: Not thread-safe. Should only be called during init.
func RegisterBuiltinPlugin(name string, plugin PluginHandler) {
	builtinPlugins[name] = plugin
	fmt.Printf("[Plugin Registry] Registered built-in plugin: %s\n", name)
}

// GetBuiltinPlugin retrieves a built-in plugin by name.
//
// Returns nil if plugin not found.
func GetBuiltinPlugin(name string) PluginHandler {
	return builtinPlugins[name]
}

// ListBuiltinPlugins returns names of all registered built-in plugins.
//
// Used by discovery system to enumerate available built-ins.
func ListBuiltinPlugins() []string {
	names := make([]string, 0, len(builtinPlugins))
	for name := range builtinPlugins {
		names = append(names, name)
	}
	return names
}
