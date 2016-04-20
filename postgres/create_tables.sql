g-- drop table if exists users, cores, readings;

create table if not exists users (
	email text primary key,
	full_name text,
	family_name text,
	given_name text,
	gender text,
	locale text,
	secret text
);

-- create table if not exists cores (
-- 	id text primary key,
-- 	owner text references users (email) on delete cascade,
-- 	name text not null,
-- 	webhookpwhash text not null
-- );

-- create table if not exists readings (
-- 	core_id text references cores (id) on delete cascade,
-- 	posted timestamp,
-- 	temp real,
-- 	humid real,
-- 	moist real,
-- 	light real,
-- 	primary key (core_id, posted)
-- );
