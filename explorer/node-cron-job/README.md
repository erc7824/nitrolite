# Node WSS Cron Job

This project is a Node.js application that queries a WebSocket server at regular intervals using a cron job. It checks the response data against a PostgreSQL database using Prisma and adds any new data that doesn't already exist in the database.

## Project Structure

```
node-wss-cron-job
├── src
│   ├── index.js          # Entry point of the application
│   ├── cron.js           # Cron job configuration
│   ├── wss-client.js      # WebSocket client management
│   ├── db
│   │   └── prisma.js      # Prisma client initialization
│   └── utils
│       └── helpers.js     # Utility functions for data operations
├── prisma
│   └── schema.prisma      # Database schema definition
├── package.json            # npm configuration file
├── .env                    # Environment variables
└── README.md               # Project documentation
```

## Installation

1. Clone the repository:
   ```
   git clone <repository-url>
   cd node-wss-cron-job
   ```

2. Install the dependencies:
   ```
   npm install
   ```

3. Set up your environment variables in the `.env` file. You will need to specify:
   - `DATABASE_URL`: Your PostgreSQL connection string.
   - `WSS_URL`: The WebSocket server URL.

## Usage

To start the application, run:
```
node src/index.js
```

This will initialize the cron job and start listening for messages from the WebSocket server.

## Cron Job Configuration

The cron job is configured to run at specified intervals. You can modify the schedule in the `src/cron.js` file.

## WebSocket Client

The WebSocket client is managed in `src/wss-client.js`. It handles the connection, sending commands, and processing incoming messages.

## Database Operations

The application uses Prisma to interact with the PostgreSQL database. The database schema is defined in `prisma/schema.prisma`. Make sure to run the migrations after setting up the schema.

## License

This project is licensed under the MIT License.