# URL Shortener

`sqlite3` `traefik` `docker`  
`vue`

Запуск проекта   
В ПАПКЕ `frontend`  
```
npm run install
npm run dev # Режим разработки 
npm run build
```

Запускам докер контейнеры
```
docker compose up --build -d --force-recreate backend-go
docker compose up --build -d --force-recreate traefik
```

Сервис будет доступен по ссылке http://app.localhost