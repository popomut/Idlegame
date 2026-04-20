# Realm of Eternity - Idle Game

A browser-based idle game built with Svelte 5 (frontend) and Go (backend).

## Project Structure

```
Idlegame/
├── frontend/     # Svelte 5 + Vite SPA
└── backend/      # Go + Fiber API server
```

## Quick Start

### Frontend (Svelte 5 + Vite)

```bash
cd frontend
npm install
npm run dev      # Start dev server at http://localhost:5173
npm run build    # Production build
npm run preview  # Preview production build
```

### Backend (Go + Fiber)

```bash
cd backend
go mod download
go run main.go   # Start server at http://localhost:3000
```

## Development

- **Frontend**: Vite dev server with HMR at `http://localhost:5173`
- **Backend**: Go API at `http://localhost:3000`
- **Database**: SQLite (auto-created on first run)

## Tech Stack

- **Frontend**: Svelte 5, Vite, Axios
- **Backend**: Go, Fiber, GORM, SQLite
- **Database**: SQLite
- **Deployment**: Railway (monorepo support)

## Deployment (Railway)

1. Push to GitHub
2. Create two services in Railway:
   - Frontend: Deploy from `/frontend`, build command: `npm run build`, start: N/A (static)
   - Backend: Deploy from `/backend`, build command: `go build`, start: `./idlegame-backend`

## Features

- **Mining System**: Extract different ore types with varying difficulty
- **Real-time Inventory**: Live ore count updates
- **Progress Tracking**: Global mining progress bar
- **Responsive Design**: Works on desktop and mobile
- **Authentication**: User login with JWT tokens

## Environment Variables

### Frontend (`.env.local`)
```
VITE_API_BASE_URL=http://localhost:3000
```

### Backend (`.env`)
```
PORT=3000
DB_PATH=idlegame.db
JWT_SECRET=popomut-secret-key-popomut99999
```

## License

MIT
