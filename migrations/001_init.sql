CREATE TABLE configurations
(
    key  text,
    value integer
);

CREATE TABLE blacklist
(
    id   serial primary key,
    ip   text,
    mask text
);

CREATE TABLE whitelist
(
    id   serial primary key,
    ip   text,
    mask text
);

INSERT INTO configurations (key, value)
VALUES ('loginAttempts', 2),
       ('passwordAttempts', 2),
       ('ipAttempts', 2);