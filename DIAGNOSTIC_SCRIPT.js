// DIAGNOSTIC SCRIPT - Run this in browser console after login

console.log('=== AUTH DIAGNOSTIC ===');
console.log('1. Checking localStorage...');
console.log('Token:', localStorage.getItem('auth_token') ? 'EXISTS' : 'MISSING');
console.log('Refresh Token:', localStorage.getItem('auth_refresh_token') ? 'EXISTS' : 'MISSING');
console.log('Username:', localStorage.getItem('auth_username'));

console.log('\n2. Testing API call...');
fetch('/api/devices', {
  headers: {
    'Authorization': 'Bearer ' + localStorage.getItem('auth_token')
  }
})
.then(r => {
  console.log('Response status:', r.status);
  return r.json();
})
.then(data => console.log('Response data:', data))
.catch(err => console.error('Error:', err));

console.log('\n3. Testing refresh token...');
fetch('/api/refresh-token', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ 
    refresh_token: localStorage.getItem('auth_refresh_token')
  })
})
.then(r => r.json())
.then(data => console.log('Refresh response:', data))
.catch(err => console.error('Refresh error:', err));
