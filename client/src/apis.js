// src/api.js

const BASE_URL = "http://localhost:8676";

export const API_ENDPOINTS = {
  leaderboard: `${BASE_URL}/leaderboard`,
  leaderboardRead: `${BASE_URL}/leaderboardRead`,
  sendPrint: `${BASE_URL}/sendPrint`,
  insertLeaderboard: `${BASE_URL}/leaderboardInsert`,
  getImage: `${BASE_URL}/getImage`
  // Add more endpoints here as needed
};
