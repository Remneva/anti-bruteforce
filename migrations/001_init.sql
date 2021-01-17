-- +goose Up
CREATE TABLE configurations
(
    key  text,
    vale text,
);

CREATE TABLE blacklist
(
    id   serial primary key,
    ip   text,
    mask text,
);

CREATE TABLE whitelist
(
    id   serial primary key,
    ip   text,
    mask text,
);
--create index black_listx on blacklist (id);
--create index white_listx on blacklist (id);

-- +goose Down
drop table configurations;
drop table cblacklist;
drop table whitelist;