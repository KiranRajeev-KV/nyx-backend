# DevOps Strategy

This document outlines the comprehensive DevOps strategy for the Nyx lost-and-found platform, including Go backend, React Admin Panel, and Flutter mobile app deployment and management.

## 🏗️ System Overview

### Architecture Components
- **Backend**: Go-based REST API with PostgreSQL database
- **Admin Panel**: React 19 + Vite web application
- **Mobile App**: Flutter application (distributed via GitHub, not Play Store)
- **Infrastructure**: Single production server deployment

### Deployment Environment
- **Development**: Local development environment only
- **Production**: Single server hosting all components
- **No staging environment**: Direct to production approach

## 🏛️ Infrastructure Setup

### Server Requirements
- **Provider**: DigitalOcean or similar cloud provider
- **Operating System**: Ubuntu LTS (latest stable)
- **Minimum Specifications**:
  - CPU: 1-2 vCPU cores
  - Memory: 2-4GB RAM
  - Storage: 10GB HDD/SSD
  - Network: Minimal bandwidth for web traffic

### Container Strategy
- **Containerization**: Docker only
- **Backend**: Multi-stage Docker image for Go application
- **Frontend**: Nginx-based containers for React/Flutter web builds
- **Database**: PostgreSQL running on host or container
- **No Kubernetes**: Simple Docker Compose setup

### Networking Architecture
```
┌─────────────────────────────────────┐
│         Production Server          │
├─────────────────────────────────────┤
│  Nginx (Reverse Proxy)         │
│  ├─ /api → Go Backend          │
│  ├─ /admin → React Admin Panel   │
│  └─ /app → Flutter Web Build    │
├─────────────────────────────────────┤
│  PostgreSQL Database             │
└─────────────────────────────────────┘
```

## 🔄 CI/CD Pipeline

### GitHub Actions Workflows

#### Backend Pipeline (Go)
```yaml
# .github/workflows/backend.yml
Trigger: Push to main, Pull requests
Steps:
1. Checkout code
2. Setup Go 1.25.5
3. Download dependencies (go mod download)
4. Run tests (go test ./...)
5. Build application (go build ./...)
6. Create Docker image
7. Deploy to production server
```

#### Frontend Pipeline (React Admin)
```yaml
# .github/workflows/admin.yml
Trigger: Push to admin branch
Steps:
1. Checkout code
2. Setup Node.js
3. Install dependencies (npm ci)
4. Build application (npm run build)
5. Deploy to server
```

#### Mobile App Pipeline (Flutter)
```yaml
# .github/workflows/flutter.yml
Trigger: Push to mobile branch
Steps:
1. Checkout code
2. Setup Flutter
3. Build web version (flutter build web)
4. Create GitHub release
5. Deploy to server
```

### Deployment Process

#### Backend Deployment
```bash
# Production deployment script
1. Pull latest code
2. Run go mod tidy
3. Run database migrations (manual approval)
4. Build new Docker image
5. Stop old container
6. Start new container
7. Health check verification
```

#### Frontend Deployment
```bash
# Admin panel deployment
1. Pull latest code
2. Install dependencies
3. Build for production
4. Copy build artifacts to nginx directory
5. Reload nginx configuration
```

## 🗄️ Database Management

### Migration Strategy
- **Approach**: Manual migrations with approval
- **Process**:
  1. Create migration files using SQLC
  2. Submit pull request for review
  3. Team approval required
  4. Manual execution during maintenance window
  5. Verify migration success

### Backup Strategy
- **Type**: Manual backups before changes
- **Schedule**: 
  - Before major deployments
  - Before database migrations
  - Before configuration changes
- **Method**: PostgreSQL dump with compression
- **Storage**: Local server + cloud storage

### Database Configuration
```sql
-- Production database settings
- Connection pooling: 20 connections
- Statement timeout: 30 seconds
- Query timeout: 60 seconds
- Log slow queries: > 500ms
```

## 🔐 Security Implementation

### Basic Security Practices
- **HTTPS**: SSL/TLS termination at reverse proxy
- **Environment Variables**: Sensitive data in .env files
- **Database Security**: Strong passwords, limited user access
- **API Security**: Input validation, rate limiting
- **Network Security**: Firewall, closed ports except HTTP(S)

### Environment Variable Management
```bash
# .env structure
DATABASE_URL=postgresql://user:pass@localhost/nyxdb
JWT_SECRET=your-super-secret-jwt-key
SMTP_HOST=smtp.gmail.com
SMTP_USER=noreply@yourdomain.com
SMTP_PASS=app-password
```

### File Security
- **No secrets in code**: All sensitive data in environment
- **.env in .gitignore**: Prevent accidental commits
- **File permissions**: Restrict access to configuration files

## 📊 Monitoring & Observability

### DIY Open-Source Monitoring Stack
```yaml
# Monitoring components
- Prometheus: Metrics collection
- Grafana: Visualization dashboards
- Logstash/ELK: Log aggregation (optional)
- Alertmanager: Alert management (optional)
```

### Application Monitoring
#### Backend (Go)
- **Logging**: Zerolog structured logging
- **Metrics**: Custom metrics endpoints
- **Health Checks**: `/health` endpoint
- **Performance**: Request timing, database queries

#### Frontend (React/Flutter)
- **Error Tracking**: Browser console errors
- **Performance**: Page load times, API response times
- **User Analytics**: Basic usage metrics

### Log Management
```bash
# Log structure and retention
- Application logs: /var/log/nyx/
- Access logs: /var/log/nginx/
- Rotation: Weekly
- Retention: 30 days
- Format: JSON structured logs
```

## 🚀 Deployment Procedures

### Production Deployment Checklist

#### Pre-Deployment
- [ ] Code reviewed and approved
- [ ] Tests passing locally
- [ ] Database backup created
- [ ] Migration scripts reviewed
- [ ] Environment variables updated
- [ ] Rollback plan prepared

#### Deployment Steps
1. **Backend Update**:
   ```bash
   cd /opt/nyx-backend
   git pull origin main
   go mod tidy
   go build ./...
   docker-compose down
   docker-compose up -d --build
   ```

2. **Frontend Update**:
   ```bash
   cd /opt/nyx-admin
   git pull origin main
   npm ci
   npm run build
   cp -r dist/* /var/www/admin/
   nginx -t && systemctl reload nginx
   ```

3. **Mobile Update**:
   ```bash
   cd /opt/nyx-mobile
   git pull origin main
   flutter build web
   cp -r build/web/* /var/www/app/
   ```

#### Post-Deployment
- [ ] Health checks passing
- [ ] Database connectivity verified
- [ ] Frontend accessible
- [ ] API endpoints responding
- [ ] Error logs reviewed
- [ ] User testing performed

### Rollback Procedures
```bash
# Backend rollback
cd /opt/nyx-backend
git checkout previous-commit
docker-compose down
docker-compose up -d

# Frontend rollback
cd /opt/nyx-admin
git checkout previous-commit
npm ci
npm run build
cp -r dist/* /var/www/admin/
systemctl reload nginx
```

## ⚡ Performance Optimization

### Backend Optimization
- **Database**: Connection pooling, query optimization
- **Caching**: Redis for frequently accessed data
- **Compression**: Gzip for API responses
- **Static Assets**: CDN integration (future)

### Frontend Optimization
- **Code Splitting**: Lazy loading for large components
- **Asset Optimization**: Image compression, minification
- **Caching**: Browser cache headers
- **Bundle Size**: Regular audits and optimization

## 🔧 Operational Management

### Server Maintenance
```bash
# Regular maintenance tasks
- System updates: Monthly
- Security patches: As needed
- Log rotation: Weekly
- Backup verification: Weekly
- Performance monitoring: Daily
```

### Service Management Commands
```bash
# Docker services
docker-compose ps          # List services
docker-compose logs -f      # View logs
docker-compose restart       # Restart services
docker-compose down        # Stop services
docker-compose up -d      # Start services

# Nginx management
nginx -t                 # Test configuration
systemctl reload nginx     # Reload configuration
systemctl restart nginx   # Restart service

# PostgreSQL management
systemctl status postgresql  # Check service
psql -d nyxdb           # Connect to database
```

## 📋 Configuration Templates

### Docker Compose (docker-compose.yml)
```yaml
version: '3.8'
services:
  nyx-backend:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=${DATABASE_URL}
    depends_on:
      - postgres
      
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: nyxdb
      POSTGRES_USER: nyxuser
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./admin_dist:/var/www/admin
      - ./mobile_dist:/var/www/app
    depends_on:
      - nyx-backend

volumes:
  postgres_data:
```

### Nginx Configuration (nginx.conf)
```nginx
events {
    worker_connections 1024;
}

http {
    upstream backend {
        server nyx-backend:8080;
    }
    
    server {
        listen 80;
        server_name your-domain.com;
        
        location /api/ {
            proxy_pass http://backend/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
        
        location /admin/ {
            root /var/www/admin;
            try_files $uri $uri/ /index.html;
        }
        
        location /app/ {
            root /var/www/app;
            try_files $uri $uri/ /index.html;
        }
    }
}
```