CREATE TABLE IF NOT EXISTS region (
    code smallint NOT NULL PRIMARY KEY,
    name varchar(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS area (
    code smallint NOT NULL PRIMARY KEY,
    region_code smallint NOT NULL references region(code),
    name varchar(255) default '',
    description varchar(255) default ''
);

CREATE TABLE IF NOT EXISTS quarter (
    code int NOT NULL PRIMARY KEY,
    area_code smallint NOT NULL references area(code),
    name varchar(255) default '',
    description varchar(255) default ''
);

CREATE TYPE object_status as ENUM ('NEW', 'SUCCESS', 'ERROR', 'NOT FOUND');

CREATE TABLE IF NOT EXISTS object (
    code int NOT NULL PRIMARY KEY,
    quarter_code int NOT NULL references quarter(code),
    load_status object_status default 'NEW',
    update_date date,
    data jsonb,

    area int,
    cost_value numeric,

    permitted_use_established_by_document varchar(128),
    right_type varchar(128),
    status varchar(128),

    land_record_type varchar(128),
    land_record_subtype varchar(128),
    land_record_category_type varchar(128)
)
