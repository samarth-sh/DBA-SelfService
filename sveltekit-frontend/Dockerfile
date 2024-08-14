# Use an official Node runtime as the base image
FROM node:20

# Set the working directory in the container
WORKDIR /app

# Copy package.json and package-lock.json to the working directory
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci

# Copy the rest of the application code
COPY . .

# Build the SvelteKit application
RUN npm run build

# Expose the port the app runs on
EXPOSE 5173

ENV PORT=5173
ENV HOST=0.0.0.0

# Command to run the application
CMD ["node", "build"]
