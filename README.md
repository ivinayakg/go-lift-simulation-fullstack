# Lift Simulation Full Stack
- A multiplayer-game, which simulates the working of lifts.

## Tech Stack
- Golang (w/ gorilla Mux)
- Mongodb
- React Native

## Setup Locally

1. Clone the project locally with
   ```
   git clone https://github.com/ivinayakg/go-lift-simulation-fullstack.git
   ```
   
2. Setup Local DB

   - You need `Docker Desktop` installed for this on your system. [check here](https://docs.docker.com/desktop/)
   - Once verified, open your `Docker Desktop` and keep it running in the background.
   - Run the following command from the repo terminal
     ```
     docker-compose -f docker-compose.yml up
     ```
   - If you want it to be running in the background add the `-d` flag.
     ```
     docker-compose -f docker-compose.yml up -d
     ```
     
3. Setup packages
   - Frontend
     ```
     cd client
     yarn
     ```
    - Backend
      ```
      cd api
      go mod tidy
      ```
4. Run Local
   - Frontend.
     ```
     cd client
     yarn dev
     ```
   - Backend
     ```
     cp api
     go run .
     ```



## Deploy

- Coming soon...


## Note

- You can reach out to me for discussion or help. (Please keep things professional and straight, don't waste my time)
- Discord: `ivinayakg`
- Email: `vinayak20029@gmail.com`
