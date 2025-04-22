# Changelog

## [2025-04-22]
### Added
- Google OAuth2 login and user creation flow.
- User is created on first OAuth login and never duplicated or overwritten on subsequent logins.
- Session handling and secure session simulation in tests.
- User API endpoints are protected; direct POST /users/ is forbidden except for future admin logic.
- Integration tests for OAuth flow, user creation, and session handling.

### Improved
- Integration tests now robustly verify that duplicate users are not created and user data is not overwritten by repeated OAuth logins.

---

*For earlier changes, see commit history or previous documentation.*
