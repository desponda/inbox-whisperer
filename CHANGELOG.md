# Changelog

## [2025-04-23]
### Changed
- Removed `migrate.tar.gz` and `test-output.txt` from git and git history.
- Improved Makefile: added `help` and `clean` targets, clarified DB migration workflow, removed `docker-migrate-up`/`down` targets for reliability.
- Updated documentation in all major markdown files to reflect new Makefile usage and DB workflow.
- Added notes about ignoring and removing large/test output files from version control.

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
