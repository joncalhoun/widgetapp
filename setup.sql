CREATE DATABASE widget_demo;
\c widget_demo

CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  token TEXT
);

CREATE TABLE widgets (
  id SERIAL PRIMARY KEY,
  userID INT NOT NULL,
  name TEXT NOT NULL,
  price INT NOT NULL,
  color TEXT NOT NULL
);

-- To speed up setup
INSERT INTO users(email) VALUES ('jon@calhoun.io');

INSERT INTO widgets(userID, name, price, color) VALUES(1, 'Go Widget', 12, 'Green');
INSERT INTO widgets(userID, name, price, color) VALUES(1, 'Slow Widget', 22, 'Yellow');
INSERT INTO widgets(userID, name, price, color) VALUES(1, 'Stop Widget', 18, 'Red');
