// Service Worker for Firebase Cloud Messaging
// File này cần được đặt ở thư mục gốc của website hoặc trong public folder
// Ví dụ: public/firebase-messaging-sw.js hoặc /firebase-messaging-sw.js

// Import Firebase scripts
importScripts('https://www.gstatic.com/firebasejs/10.7.1/firebase-app-compat.js');
importScripts('https://www.gstatic.com/firebasejs/10.7.1/firebase-messaging-compat.js');

// Firebase config - CẦN THAY ĐỔI THEO PROJECT CỦA BẠN
const firebaseConfig = {
  apiKey: "AIzaSyASJbpleMfK4LC6hGIhcJC0ng1GtkFi4yY",
  authDomain: "api--core.firebaseapp.com",
  projectId: "api--core",
  storageBucket: "api--core.firebasestorage.app",
  messagingSenderId: "295353376871",
  appId: "1:295353376871:web:7f837bad391e4278bc34c2",
  measurementId: "G-4MT0VSYBFW"
};

// Initialize Firebase
firebase.initializeApp(firebaseConfig);

// Retrieve Firebase Messaging object
const messaging = firebase.messaging();

// Handle background messages
messaging.onBackgroundMessage((payload) => {
  console.log('[firebase-messaging-sw.js] Received background message ', payload);

  // Customize notification here
  const notificationTitle = payload.notification?.title || 'New Notification';
  const notificationOptions = {
    body: payload.notification?.body || 'You have a new notification',
    icon: payload.notification?.icon || '/favicon.ico',
    badge: '/favicon.ico',
    tag: payload.data?.tag || 'fcm-notification',
    data: payload.data || {},
    requireInteraction: false,
    silent: false,
  };

  // Show notification
  return self.registration.showNotification(notificationTitle, notificationOptions);
});

// Handle notification click
self.addEventListener('notificationclick', (event) => {
  console.log('[firebase-messaging-sw.js] Notification click received.');

  event.notification.close();

  // Open or focus the app when notification is clicked
  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
      // If app is already open, focus it
      for (const client of clientList) {
        if (client.url === '/' && 'focus' in client) {
          return client.focus();
        }
      }
      // Otherwise open a new window
      if (clients.openWindow) {
        return clients.openWindow('/');
      }
    })
  );
});

// Handle notification close
self.addEventListener('notificationclose', (event) => {
  console.log('[firebase-messaging-sw.js] Notification closed:', event.notification);
});

