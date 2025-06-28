// Simple test script for movie club functionality
const axios = require('axios');

const baseURL = 'http://localhost:3080/api';
let authToken = '';

async function login() {
  try {
    // You'll need to replace with actual login credentials
    const response = await axios.post(`${baseURL}/auth/login`, {
      username: 'testuser',
      password: 'testpass'
    });
    authToken = response.data.token;
    console.log('✅ Login successful');
    return true;
  } catch (error) {
    console.log('❌ Login failed:', error.response?.data?.error || error.message);
    return false;
  }
}

async function testMovieClub() {
  const headers = { Authorization: `Bearer ${authToken}` };
  
  try {
    // Test getting current cycle
    console.log('Testing current cycle endpoint...');
    const cycleResponse = await axios.get(`${baseURL}/movie-club/current`, { headers });
    console.log('✅ Current cycle:', cycleResponse.data.cycle?.phase || 'No active cycle');
    
    // Test nomination (if in nomination phase)
    if (cycleResponse.data.cycle?.phase === 'nomination' && cycleResponse.data.canNominate) {
      console.log('Testing nomination...');
      const nominateResponse = await axios.post(`${baseURL}/movie-club/nominate`, {
        contentId: 550, // Fight Club TMDB ID
        reason: 'Classic movie test'
      }, { headers });
      console.log('✅ Nomination successful');
    }
    
    // Test voting (if in voting phase)
    if (cycleResponse.data.cycle?.phase === 'voting' && cycleResponse.data.canVote) {
      console.log('Testing voting...');
      const voteResponse = await axios.post(`${baseURL}/movie-club/vote`, {
        votes: [
          { contentId: 550, priority: 1 },
          { contentId: 680, priority: 2 }
        ]
      }, { headers });
      console.log('✅ Voting successful');
    }
    
    // Test results (if in watching phase)
    if (cycleResponse.data.cycle?.phase === 'watching') {
      console.log('Testing results...');
      const resultsResponse = await axios.get(`${baseURL}/movie-club/results`, { headers });
      console.log('✅ Results retrieved:', resultsResponse.data.length, 'movies');
    }
    
  } catch (error) {
    console.log('❌ Test failed:', error.response?.data?.error || error.message);
  }
}

async function runTests() {
  console.log('🎬 Starting Movie Club Tests\n');
  
  const loginSuccess = await login();
  if (!loginSuccess) {
    console.log('Cannot proceed without authentication');
    return;
  }
  
  await testMovieClub();
  console.log('\n🏁 Tests completed');
}

runTests();