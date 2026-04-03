# build the images
docker build -f Dockerfile-Zoro -t zoro-backend ./backend
docker build -f Dockerfile-Sanji -t sanji-genai ./genAI
docker build -f Dockerfile-Nami -t nami-frontend ./frontend

# create a docker network
docker network create strawhats

# start the containers
docker run -d --env-file .env --network strawhats --name sanji -p 4000:4000 sanji-genai
docker run -d --env-file .env --network strawhats --name zoro -p 3000:3000 zoro-backend
docker run -d --network strawhats --name nami -p 8080:80 nami-frontend

# health checks
curl http://localhost:3000/health
curl http://localhost:4000/health
curl http://localhost:8080