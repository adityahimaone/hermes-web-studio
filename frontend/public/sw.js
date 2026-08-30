const VERSION = 'hermes-studio-shell-v1'
const SHELL = [
  './',
  './index.html',
  './manifest.webmanifest',
  './icons/hermes-studio.svg',
]

self.addEventListener('install', event => {
  event.waitUntil(caches.open(VERSION).then(cache => cache.addAll(SHELL)).then(() => self.skipWaiting()))
})

self.addEventListener('activate', event => {
  event.waitUntil(caches.keys().then(keys => Promise.all(keys.filter(key => key !== VERSION).map(key => caches.delete(key)))).then(() => self.clients.claim()))
})

self.addEventListener('fetch', event => {
  if (event.request.method !== 'GET' || new URL(event.request.url).origin !== self.location.origin) return
  if (new URL(event.request.url).pathname.includes('/api/')) return
  event.respondWith(fetch(event.request).then(response => {
    if (response.ok && event.request.destination !== 'document') {
      const copy = response.clone()
      void caches.open(VERSION).then(cache => cache.put(event.request, copy))
    }
    return response
  }).catch(() => caches.match(event.request).then(response => response || caches.match('./index.html'))))
})
