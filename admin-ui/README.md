# Edge.link Admin UI

A modern SaaS admin interface for managing your Edge.link Proxy-as-a-Service configuration built with Next.js, Tailwind CSS, and PostgreSQL.

## Features

### 🎛️ **Complete Configuration Management**
- **Route Management**: CRUD operations for proxy routes with full configuration options
- **API Key Management**: Generate, view, edit, and revoke API keys with permissions
- **Real-time Dashboard**: Live metrics, statistics, and system health monitoring
- **Activity Logs**: Recent activity tracking and request logs

### 🎨 **Modern UI/UX**
- **Responsive Design**: Mobile-first design that works on all devices
- **Tailwind CSS**: Beautiful, consistent styling with dark/light themes
- **Interactive Components**: Form validation, modals, and real-time updates
- **Accessibility**: WCAG compliant with keyboard navigation and screen reader support

### 🔐 **Authentication & Security**
- **User Authentication**: Secure login system with session management
- **Role-based Access**: Admin and user roles with permission management
- **API Security**: JWT tokens and CSRF protection
- **Data Validation**: Client and server-side validation with Zod schemas

### 📊 **Analytics & Monitoring**
- **Real-time Metrics**: Request counts, response times, error rates
- **Route Analytics**: Per-route performance statistics and cache hit rates
- **Visual Charts**: Interactive graphs and charts for data visualization
- **Log Management**: Searchable and filterable request logs

## Quick Start

### Prerequisites
- Node.js 18+ 
- PostgreSQL 12+
- Your Edge.link proxy service running

### 1. Database Setup

```bash
# Create PostgreSQL database
createdb edgelink

# Run the schema
psql -d edgelink -f schema.sql
```

### 2. Install Dependencies

```bash
cd admin-ui
npm install
```

### 3. Environment Configuration

Create `.env.local`:

```bash
# Database
POSTGRES_URL=postgresql://username:password@localhost:5432/edgelink

# NextAuth.js
NEXTAUTH_SECRET=your-secret-key-here
NEXTAUTH_URL=http://localhost:3000

# Edge.link Proxy API
PROXY_API_URL=http://localhost:8080
```

### 4. Start Development Server

```bash
npm run dev
```

The admin UI will be available at `http://localhost:3000`.

## Default Login

- **Email**: `admin@example.com`
- **Password**: `admin123`

## Project Structure

```
admin-ui/
├── pages/                  # Next.js pages and API routes
│   ├── api/               # Backend API endpoints
│   │   ├── routes/        # Route management API
│   │   ├── keys/          # API key management API
│   │   └── dashboard/     # Dashboard data API
│   ├── routes/            # Route management pages
│   ├── keys/              # API key management pages
│   └── index.tsx          # Dashboard page
├── components/            # Reusable React components
│   ├── layouts/           # Layout components
│   ├── forms/             # Form components
│   └── ui/                # UI components
├── lib/                   # Utility libraries
│   └── db.ts              # Database connection and models
├── styles/                # CSS styles
└── public/                # Static assets
```

## Key Features

### Dashboard
- **System Overview**: Total requests, response times, success rates
- **Quick Actions**: Fast access to common tasks
- **Recent Activity**: Live feed of system events
- **Health Monitoring**: Service status and uptime tracking

### Route Management
- **Visual Route Builder**: Drag-and-drop route configuration
- **Advanced Options**: Cache settings, rate limits, authentication
- **Validation**: JSON schema validation configuration
- **Testing**: Built-in route testing and debugging tools

### API Key Management
- **Secure Generation**: Cryptographically secure key generation
- **Permission System**: Granular permission control with wildcards
- **Usage Analytics**: Per-key usage statistics and rate limiting
- **Expiration Management**: Automatic key expiration and renewal

### Analytics & Logs
- **Performance Metrics**: Detailed performance analysis
- **Error Tracking**: Error rates and debugging information
- **Cache Analytics**: Cache hit rates and performance optimization
- **Request Logs**: Detailed request logging with filtering

## API Integration

The admin UI communicates with your Edge.link proxy service through REST APIs:

### Database Schema
The UI stores configuration in PostgreSQL with these main tables:
- `users` - Admin user accounts
- `proxy_routes` - Route configurations
- `api_keys` - API key management
- `request_logs` - Request logging and analytics

### Configuration Sync
The UI provides REST endpoints that your proxy service can call to fetch the latest configuration:

```bash
# Get all routes
GET /api/routes

# Get all API keys  
GET /api/keys

# Get specific route
GET /api/routes/{id}
```

## Development

### Adding New Features

1. **Create Database Schema**: Add tables to `schema.sql`
2. **Add Database Functions**: Extend `lib/db.ts` with new operations
3. **Create API Endpoints**: Add routes to `pages/api/`
4. **Build UI Components**: Create reusable components in `components/`
5. **Add Pages**: Create new pages in `pages/`

### Form Validation

All forms use React Hook Form with Zod validation:

```typescript
const schema = z.object({
  path: z.string().min(1, 'Path is required'),
  target: z.string().url('Must be a valid URL'),
});
```

### Database Operations

Database operations use the centralized DB class:

```typescript
// Create new route
const route = await DB.createRoute({
  path: '/api/v1/users',
  target: 'https://api.example.com',
  methods: ['GET', 'POST'],
  cache_enabled: true,
});
```

## Deployment

### Production Build

```bash
npm run build
npm start
```

### Docker Deployment

```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build
CMD ["npm", "start"]
```

### Environment Variables

Required for production:
- `POSTGRES_URL` - PostgreSQL connection string
- `NEXTAUTH_SECRET` - Secret for session encryption
- `NEXTAUTH_URL` - Full URL of your deployment
- `PROXY_API_URL` - URL of your Edge.link proxy service

## Security

### Authentication
- NextAuth.js for secure session management
- Password hashing with bcrypt
- CSRF protection on all forms

### Authorization
- Role-based access control
- API endpoint protection
- Resource-level permissions

### Data Protection
- SQL injection prevention with parameterized queries
- XSS protection with React's built-in escaping
- Secure headers and Content Security Policy

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/new-feature`
3. Make your changes and add tests
4. Commit your changes: `git commit -m 'Add new feature'`
5. Push to the branch: `git push origin feature/new-feature`
6. Submit a pull request

## License

This project is licensed under the MIT License - see the main project LICENSE file for details.