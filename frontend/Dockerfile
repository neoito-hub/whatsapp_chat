# Use the official Node.js runtime as the base image for building the app
FROM node:20-alpine as build

# Install Python, make, and g++
RUN apk add --no-cache python3 make g++

# Set the working directory in the container
WORKDIR /app

# Install pnpm globally
RUN npm install -g pnpm

# Copy package.json and pnpm-lock.yaml to the working directory
COPY package*.json ./

# Install dependencies with pnpm
RUN pnpm install

# Copy the entire application code to the container
COPY . .

# Build the React app for production
RUN pnpm run build

# Debugging step: List the contents of /app directory to verify build output
RUN ls -la /app

# Use Nginx as the production server
FROM nginx:alpine

# Copy the built React app to Nginx's web server directory
COPY --from=build /app/dist /usr/share/nginx/html

# Expose port 80 for the Nginx server
EXPOSE 80

# Start Nginx when the container runs
CMD ["nginx", "-g", "daemon off;"]
