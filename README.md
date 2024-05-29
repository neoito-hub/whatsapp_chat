# WhatsApp Chat

## Overview

WhatsApp Chat is a versatile tool that enables seamless communication with users via WhatsApp. With this chat, you can send messages, craft templates, engage in one-on-one conversations, and broadcast messages efficiently.

## Components

1. **Backend**: The backend component contains APIs responsible for integrating with WhatsApp, handling message processing, user interactions, and data management.

2. **Authentication Shield**: Ensures secure access to the chat system by implementing robust authentication mechanisms to protect sensitive user data and system functionalities.

3. **Frontend Application**: The frontend application provides users with a user-friendly interface to interact with your system. It communicates with the backend components, leveraging the authentication mechanism provided by Shield.

## Prerequisites

Before getting started, ensure the following prerequisites are met:

- Docker installed on your machine
- BB CLI (BB Command Line Interface) installed

## Getting Started

1. **Clone Repository**: Begin by cloning the repository to your local machine.
   
2. **Replace Meta Credentials**: In the provided seed file in shared folder, replace the placeholder credentials with your actual meta credentials in the project table.

3. **Build Docker Compose**: Execute the following commands in your terminal:
    ```bash
    docker compose build
    ```

4. **Start the Project**: Once the build process is complete, start the project with:
    ```bash
    docker compose up
    ```
## License

This project is licensed under the [MIT License](LICENSE).
