# Modern SaaS UI Template (React + Tailwind + DaisyUI)

This template replicates the exact Inbox Whisperer UI setup—perfect for new projects that want a beautiful, modern, dark-mode SaaS landing page and design system.

---

## Features
- **Shiny dark gradient background** (radial/linear, navy/blue/cyan)
- **DM Sans font** for a modern, geometric look
- **Centered SaaS hero section** with bold headline, accent gradient, and CTA
- **Accent color highlights** (e.g., #14e0c9)
- **Responsive, minimal navbar and footer**
- **No glassmorphism or excessive shadows**
- **Tailwind CSS + DaisyUI** for rapid UI building

---

## Folder Structure Best Practices
- Use `web` or `ui` or `frontend` for your main React app (e.g., `web/` or `frontend/`).
- Avoid legacy names like `mvp-ui`—always use `web` for clarity and convention.
- Place all React code, public assets, and fonts in this directory.

---

## Replication Steps
1. **Copy the folder** (e.g., `web/`) to your new project.
2. **Install dependencies:**
   ```bash
   npm install
   # or
yarn install
   ```
3. **Start the dev server:**
   ```bash
   npm run dev
   # or
yarn dev
   ```
4. **Fonts:**
   - Ensure `web/public/fonts/dmsans.css` is present and imported in your main layout (e.g., `src/pages/_app.tsx` or `index.tsx`).
   - DM Sans is loaded from Google Fonts in that CSS file.
5. **Theme:**
   - Use the shiny dark gradient background as in the Home page.
   - Reference the Home page for layout, accent color, and spacing.
6. **.gitignore:**
   - Ensure `node_modules/`, `build/`, `dist/`, `.env`, font cache, and local dev files are ignored.

---

## Design System
- **Font:** DM Sans everywhere
- **Background:** Shiny dark gradient (see Home page CSS)
- **Accent:** #14e0c9 for buttons, highlights, gradients
- **Layout:** Centered, bold, clean, responsive

---

## Example Home Page (JSX)
```jsx
// See your Inbox Whisperer Home.tsx for the full example
```

---

## Credits
- Inspired by top-tier SaaS landing pages (Thunderbird, Linear, Vercel, Superhuman)
- Inbox Whisperer UI by [your team]
