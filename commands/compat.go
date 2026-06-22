package commands

// Top-level login/logout aliases mirror the official @n8n/cli command surface.
// They reuse the same logic as `auth login` / `auth logout` (fresh command
// instances; cobra does not allow one command under two parents).
func init() {
	login := authLoginCmd()
	login.Use = "login"
	login.Short = "Authenticate the active profile (alias for `auth login`)"
	rootCmd.AddCommand(login)

	logout := authLogoutCmd()
	logout.Use = "logout"
	logout.Short = "Remove the active profile's API key (alias for `auth logout`)"
	rootCmd.AddCommand(logout)
}
