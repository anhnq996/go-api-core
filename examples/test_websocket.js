// WebSocket Test Script
const WebSocket = require('ws');

console.log('🚀 WebSocket Test Script - ApiCore');
console.log('===================================');

// Create WebSocket connection
const ws = new WebSocket('ws://localhost:3000/ws?user_id=test_user_123');

ws.on('open', function () {
  console.log('✅ Connected to WebSocket server');

  // Test 1: Join a room
  console.log('\n🧪 Test 1: Join Room');
  ws.send(JSON.stringify({
    type: 'join_room',
    data: 'general'
  }));

  // Test 2: Send a message to room
  setTimeout(() => {
    console.log('\n🧪 Test 2: Send Room Message');
    ws.send(JSON.stringify({
      type: 'room_message',
      data: {
        room: 'general',
        message: 'Hello from test script!'
      }
    }));
  }, 1000);

  // Test 3: Send broadcast message
  setTimeout(() => {
    console.log('\n🧪 Test 3: Send Broadcast Message');
    ws.send(JSON.stringify({
      type: 'broadcast',
      data: 'Broadcast message from test script!'
    }));
  }, 2000);

  // Test 4: Send notification
  setTimeout(() => {
    console.log('\n🧪 Test 4: Send Notification');
    ws.send(JSON.stringify({
      type: 'notification',
      data: {
        title: 'Test Notification',
        body: 'This is a test notification from WebSocket test script',
        type: 'info'
      }
    }));
  }, 3000);

  // Test 5: Leave room
  setTimeout(() => {
    console.log('\n🧪 Test 5: Leave Room');
    ws.send(JSON.stringify({
      type: 'leave_room',
      data: 'general'
    }));
  }, 4000);

  // Close connection after tests
  setTimeout(() => {
    console.log('\n✅ All tests completed, closing connection...');
    ws.close();
  }, 5000);
});

ws.on('message', function (data) {
  const message = JSON.parse(data.toString());
  console.log(`📨 Received: ${JSON.stringify(message, null, 2)}`);
});

ws.on('close', function () {
  console.log('❌ Disconnected from WebSocket server');
  process.exit(0);
});

ws.on('error', function (error) {
  console.error('❌ WebSocket error:', error);
  process.exit(1);
});
