const WebSocket = require('ws');

function connectToWSS(url, message, onMessageCallback) {
  const ws = new WebSocket(url);

  ws.on('open', () => {
    console.log('Connected to WSS');
    ws.send(JSON.stringify(message));
    console.log('Message sent to WSS:', message);
  });

  ws.on('message', (data) => {
    console.log('Message received from WSS:', data);
    const parsedData = JSON.parse(data);
    onMessageCallback(parsedData);
  });

  ws.on('error', (error) => {
    console.error('WebSocket error:', error);
  });

  ws.on('close', () => {
    console.log('WebSocket connection closed');
  });
}

module.exports = { connectToWSS };