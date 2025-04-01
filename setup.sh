#!/bin/bash

# Create PostgreSQL database and user
psql postgres -c "CREATE USER voicethread WITH PASSWORD 'voicethread';"
psql postgres -c "CREATE DATABASE voicethread_development;"
psql postgres -c "GRANT ALL PRIVILEGES ON DATABASE voicethread_development TO voicethread;"

# Install Rails dependencies
cd voicethread
bundle install

# Run Rails migrations
bundle exec rails db:migrate

# Install Go dependencies
cd ../backend
go mod tidy

# Install frontend dependencies
cd ../frontend
npm install

# Create storage directory
cd ../backend
mkdir -p storage

echo "Development environment setup complete!"
echo "To start the services:"
echo "1. Start PostgreSQL"
echo "2. Start Rails: cd voicethread && rails s"
echo "3. Start Go backend: cd backend && go run cmd/server/main.go"
echo "4. Start frontend: cd frontend && npm run dev" 