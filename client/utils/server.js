import axios from "axios";

const baseurl = "https://lift-api.ivinayakg.me";

const fetch = axios.create({
  baseURL: baseurl,
});

const createSession = async (floors, lifts) => {
  const response = await fetch.post(`/session`, { floors: 9, lifts: 4 });
  return response.data;
};

const fetchSession = async (sessionId) => {
  const response = await fetch.get(`/session/${sessionId}`);
  return response.data;
};

const createRequest = async (sessionId, clientId, floor) => {
  const response = await fetch.post(`/session/${sessionId}/request`, {
    floor,
    clientId,
  });
  return response.data;
};

export { createSession, fetchSession, createRequest, baseurl };
