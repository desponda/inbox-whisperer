# Developing.md

## Project Overview

Inbox Whisperer is a suite of AI-powered tools designed to help users achieve and maintain inbox zero. Our first feature is the Inbox Zero Helper—a guided workflow that enables users to efficiently triage and organize their Gmail inbox, ensuring that only truly important emails remain.

## MVP Focus (2025-04-22)
- Users sign up/log in via Google OAuth2
- After login, users can fetch their emails (list)
- Users can view the content of a specific email
- All endpoints are described in OpenAPI and the spec is always kept up-to-date
- **Frontend is being built as a React SPA, using strict best practices, a UI framework, and an auto-generated OpenAPI client.**

## Core Workflow

1. **Email Fetching**: The app securely connects to Gmail and fetches the user's emails.

### Email Fetching Best Practices
- **Inbox loads instantly from DB cache:** When a user loads the inbox, the backend returns cached email summaries from the database for a fast, snappy experience.
- **Background sync:** Simultaneously, the backend triggers a background sync with Gmail to fetch new/updated summaries (using the Gmail API's list/metadata endpoints). The cache is updated in the background.
- **Summaries vs. Full Content:** The summary endpoint (`FetchMessages`) returns only minimal fields (subject, sender, snippet, date, etc.)—never the full body/content. The full content is fetched only via a separate endpoint (`FetchMessageContent`) and only for the selected message.
- **Client updates:** After sync, the client can poll or receive a push notification to refresh the inbox view with new data.

2. **Categorization**: Each email is automatically categorized using AI into one of four actionable groups:
   - **Promotions/Ads (FYI)**: Marketing emails, advertisements, and newsletters—content the user likely doesn't want to open or act on.
   - **To Review**: Emails the user probably wants to read, but that do not require a response (e.g., updates, reports, notifications).
   - **Important**: Emails that are personal, time-sensitive, or require a response or action.
   - **Deferred**: Emails that require action or follow-up at a later time.

   (Note: These categories are designed to be intuitive and actionable. If you have suggestions for more effective categories, please document them here.)

3. **Guided Cleanup**:
   - Users first review "Promotions/Ads (FYI)" emails, with the option to bulk archive.
   - Next, users process "To Review" emails, choosing to archive, defer, or mark as "Important".
   - Then, users review "Deferred" emails to schedule follow-up actions.
   - Finally, users review "Important" emails to confirm their status or take necessary action.

4. **Desired Outcome**: The goal is for users to either archive or appropriately categorize every email, leaving only genuinely important or deferred items in their inbox.

## Design and Coding Principles

### Backend
- **RESTful API**: The backend will be designed as a strongly RESTful API using the go-chi framework for routing and middleware.
- **Separation of Concerns**: Code will be organized to ensure clear separation between business logic, data access, and presentation layers.
- **Clean Code**: We will follow clean coding principles—readable, maintainable, and well-documented code.
- **Interfaces & Abstraction**: Use interfaces and abstract patterns where appropriate to enable flexibility, testing, and future extension.

### Frontend
- **React**: The frontend will be built with React, focusing on a responsive and intuitive user experience.

### API Contract
- **OpenAPI Spec**: All API endpoints will be defined with OpenAPI. This ensures a strong, always-updated contract between frontend and backend, and enables auto-generation of the JavaScript client for React.
- **Auto-Generated Clients**: The OpenAPI spec will be used to generate and update the frontend API client, reducing manual work and preventing contract drift.

## Feature Development Workflow

- Each feature will have its own Markdown file within a dedicated `features/` directory.
- Each feature file will include:
  - An implementation plan.
  - A checklist of tasks for AI tools to read, update, and track progress.
- Workflow for new features:
  1. The AI will first create an implementation plan for the feature.
  2. The AI will generate a checklist of actionable tasks.
  3. The AI will proceed with development, updating the checklist as progress is made.
  4. After each development session, changes will be committed to checkpoint progress and maintain context for future sessions.
- This approach ensures persistent, up-to-date context for both humans and AI tools, supporting efficient, incremental development.

## Development Notes

- This document is maintained by the Windsurf context tracker.
- Please update with any new insights, workflow changes, or category refinements as development progresses.
