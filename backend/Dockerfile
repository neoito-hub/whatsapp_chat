
FROM node:18.12

WORKDIR /app

COPY ./shared/package.json ./ 
COPY ./shared ./

COPY . .

RUN  npm i @appblocks/bb-cli -g
RUN npm i prisma
RUN npm i pm2 -g

EXPOSE 5001 5000 5002

CMD ["bb", "start"]
