-- +goose Up
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
--create index black_listx on blacklist (id);
--create index white_listx on blacklist (id);

INSERT INTO configurations (key, value)
VALUES ('loginAttempts', 10),
       ('passwordAttempts', 100),
       ('ipAttempts', 1000);

-- +goose Down
drop table configurations;
drop table cblacklist;
drop table whitelist;

