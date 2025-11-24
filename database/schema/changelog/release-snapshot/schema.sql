CREATE TABLE IF NOT EXISTS region (
    code smallint NOT NULL PRIMARY KEY,
    name varchar(255) NOT NULL,
    description varchar(255) default ''
);

CREATE TABLE IF NOT EXISTS area (
    code smallint NOT NULL PRIMARY KEY,
    region_code smallint NOT NULL references region(code),
    name varchar(255) NOT NULL,
    description varchar(255) default ''
);

CREATE TABLE IF NOT EXISTS quarter (
    code smallint NOT NULL PRIMARY KEY,
    area_code smallint NOT NULL references area(code),
    name varchar(255) NOT NULL,
    description varchar(255) default ''
);

CREATE TABLE IF NOT EXISTS object (
    code smallint NOT NULL PRIMARY KEY,
    quarter_code smallint NOT NULL references quarter(code),
    data jsonb NOT NULL
)
