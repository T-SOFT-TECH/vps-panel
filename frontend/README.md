# VPS Panel - Frontend

Modern SvelteKit frontend for the VPS Panel deployment platform.

## 🎨 Tech Stack

- **SvelteKit** - Latest with Svelte 5 (runes)
- **Tailwind CSS v4** - Utility-first styling
- **TypeScript** - Type safety
- **Vite** - Fast development & build

## 🚀 Getting Started

### Install Dependencies

```bash
npm install
```

### Configure Environment

```bash
cp .env .env
```

Update the `.env` file with your API URL (defaults to `http://localhost:8080/api/v1`).

### Run Development Server

```bash
npm run dev
```

The frontend will be available at http://localhost:5173

### Build for Production

```bash
npm run build
```

### Preview Production Build

```bash
npm run preview
```

## 📁 Project Structure

```
src/
├── lib/
│   ├── api/                 # API clients
│   │   ├── client.ts        # Base API client
│   │   ├── auth.ts          # Auth endpoints
│   │   ├── projects.ts      # Project endpoints
│   │   └── deployments.ts   # Deployment endpoints
│   ├── components/          # Reusable UI components
│   │   ├── Button.svelte
│   │   ├── Card.svelte
│   │   ├── Input.svelte
│   │   ├── Select.svelte
│   │   ├── Badge.svelte
│   │   ├── Modal.svelte
│   │   ├── Alert.svelte
│   │   └── Navbar.svelte
│   ├── stores/              # Svelte stores (using runes)
│   │   └── auth.svelte.ts   # Auth state management
│   ├── types.ts             # TypeScript type definitions
│   └── utils/               # Utility functions
│       └── format.ts        # Formatting helpers
│
├── routes/                  # SvelteKit pages
│   ├── +layout.svelte       # Root layout
│   ├── +page.svelte         # Home (redirect)
│   ├── login/               # Login page
│   ├── register/            # Register page
│   └── (app)/               # Protected app routes
│       ├── +layout.svelte   # App layout with navbar
│       ├── +page.svelte     # Dashboard
│       └── projects/        # Project pages
│           ├── +page.svelte              # Projects list
│           ├── new/+page.svelte          # New project form
│           └── [id]/                     # Project detail
│               ├── +page.svelte          # Project overview
│               └── deployments/
│                   └── [deploymentId]/+page.svelte  # Deployment logs
│
├── app.css                  # Tailwind CSS imports
└── app.html                 # HTML template
```

## 🎯 Features

### Authentication
- Login & Registration
- JWT token management
- Auto-redirect based on auth state
- Secure token storage in localStorage

### Dashboard
- Project statistics
- Recent deployments
- Quick actions

### Projects
- List all projects
- Create new project with form
- Project detail view
- Delete project
- Search/filter projects

### Deployments
- View deployment history
- Real-time deployment logs
- Deployment status tracking
- Cancel in-progress deployments

### UI Components
All components built with Svelte 5 runes for reactivity:
- Button (with loading states)
- Card
- Input (with validation)
- Select dropdown
- Badge (status indicators)
- Modal dialogs
- Alert messages
- Navbar

## 🎨 Tailwind CSS v4

This project uses **Tailwind CSS v4**, which has a different configuration approach:

### Import in CSS
```css
/* src/app.css */
@import 'tailwindcss';
@plugin '@tailwindcss/forms';
@plugin '@tailwindcss/typography';
```

### No tailwind.config.js
Tailwind v4 doesn't use a config file - styling is done directly in CSS with `@theme` directive.

### Vite Plugin
```js
// vite.config.ts
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [tailwindcss(), sveltekit()]
});
```

## 🔧 Development

### Type Checking

```bash
npm run check
```

### Formatting

```bash
npm run format
```

### Linting

```bash
npm run lint
```

## 🏗️ Building

### Docker Build

```bash
docker build -t vps-panel-frontend .
docker run -p 3000:3000 vps-panel-frontend
```

### Environment Variables

- `VITE_API_URL` - Backend API URL (default: http://localhost:8080/api/v1)

## 📱 Responsive Design

The UI is fully responsive and works on:
- Desktop (1024px+)
- Tablet (768px - 1023px)
- Mobile (< 768px)

## 🎭 State Management

Using Svelte 5 runes for reactive state:

```typescript
// Example: Auth store
class AuthStore {
  user = $state<User | null>(null);
  token = $state<string | null>(null);

  get isAuthenticated() {
    return !!this.token && !!this.user;
  }
}
```

## 🔐 API Integration

All API calls go through the centralized client:

```typescript
import { api } from '$lib/api/client';

// Automatically includes auth token
const projects = await api.get('/projects');
```

## 🚦 Routing

Using SvelteKit's file-based routing with route groups:

- `/` - Redirect to dashboard or login
- `/login` - Public login page
- `/register` - Public registration page
- `/(app)/*` - Protected routes (requires auth)

## 📄 License

MIT License - see root LICENSE file
