package storage

const queryAddAnimal = `insert into animal_store (animal_id, animal, price) values ($1, $2, $3)`
