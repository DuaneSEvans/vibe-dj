# Goals

The goal of this project is to be able to play a computer game and have AI
generate the right song based on the ambience and mood of the game being played.
It should then automatically play the song via spotify.

To do this, a frontend app, maybe a google chrome extension, takes a screenshot
of the activity, then sends it to the server. The server interprets the image
using an LLM and describes the
image. The description returned in optimized and formatted such that it can be
given to Spotify's API and get a song returned. Then the song is sent to the
front end and is played. Ideally, the song is played automatically possibly in
a google chrome extension or on whatever device is already playing music.
Ideally though, the computer that took the screenshot should start playing the
music.

The project should work with agentic tooling and some devops deploying the app
to AWS. Also, it is a project to teach me more GoLang.

See TODO.md file for a breakdown of tasks.

# Project TODO List: AI Game DJ (Client-Server Edition)

This plan reflects a client-server architecture with a browser extension frontend and a Go backend on AWS.

### Frontend (Browser Extension)

- [ ] **1. Scaffold Extension:** Scaffold a basic browser extension (e.g., for Chrome).
- [ ] **2. Screenshot:** Implement screenshot functionality in the extension.
- [ ] **3. Spotify Auth:** Implement the Spotify OAuth flow in the extension to get a user token.
- [ ] **4. API Call:** Make the extension send the screenshot and auth token to the backend API.
- [ ] **5. Display Song:** Handle the response from the backend to display/play the recommended song.

### Backend (Go Service on AWS)

- [ ] **1. API Endpoint:** Create a Go server with an API endpoint to receive image data and a Spotify token.
- [ ] **2. Host LLM:** Research and choose a platform for hosting the LLaVA model (e.g., Replicate, SageMaker).
- [ ] **3. Call LLM:** Implement backend logic to call the hosted LLM API with the image and prompt.
- [ ] **4. Call Spotify:** Use the user's Spotify token from the request to call the Spotify API for a track.
- [ ] **5. Deploy:** Containerize and deploy the Go service to a suitable AWS service (e.g., App Runner or Fargate).
