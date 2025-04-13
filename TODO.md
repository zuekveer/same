# API

0. golang, postman, postgres, docker
1. Сделать ручки, пиши в одном файле main.go, данные храним в глобальной переменной (map, слайс)
    Используйем fiber
    1. `POST` `/user` request body{name, age, ...}, resp id, 201, 400, 500
    2. `PUT` `/user` request body{id, name, age, ...}, resp id, 200, 400, 500, 404
    3. `GET` `/user/:id` resp body{name, age, ...}, 200, 400, 500, 404
    4. `DELETE` `/user/:id` 200, 400, 500, 404

1. слайс vs массив, структура слайса, что происходить при append (cap)
2. map, бакеты, эвакуации, коллизии, чек про swiss table
3. interface, что под капотом, solid, опп есть или нет (как реализуется в гошке)
4. Приведение типов

