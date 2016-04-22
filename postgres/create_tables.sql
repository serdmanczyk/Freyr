-- drop table if exists users, cores, readings;

create table if not exists users (
	email text primary key,
	full_name text,
	family_name text,
	given_name text,
	gender text,
	locale text,
	secret text
);

create table if not exists readings (
    useremail text references users(email),
    posted timestamp,
    coreid text,
    temperature real,
    humidity real,
    moisture real,
    light real,
    battery real,
    primary key (useremail, coreid, posted)
);


-- insert into users (email, full_name, family_name, given_name, gender) values
--     ('serdmanczyk@gmail.com', 'steven erdmanczyk', 'erdmanczyk', 'steven', 'mail'),
--     ('sjohn@gmail.com', 'steven john', 'john', 'steven', 'male'),
--     ('sajohn@gmail.com', 'sarah john', 'john', 'sarah', 'female'),
--     ('kjo@gmail.com', 'kim jo', 'jo', 'kim', 'female');

-- insert into readings (useremail, posted, coreid, temperature, humidity, moisture, light, battery) values
--     ('serdmanczyk@gmail.com', '2015-12-31T12:12:12', '0987901287340918723409', '18.00', '40.50', '50.0', '114.00', '50.0'),
--     ('serdmanczyk@gmail.com', '2015-12-31T12:12:14', '0987901287340918723409', '18.10', '40.40', '50.1', '114.00', '50.0'),
--     ('serdmanczyk@gmail.com', '2015-12-31T12:12:16', '0987901287340918723409', '18.20', '40.30', '50.2', '114.00', '50.0'),
--     ('serdmanczyk@gmail.com', '2015-12-31T12:12:18', '0987901287340918723409', '18.10', '40.40', '50.0', '114.00', '50.0'),
--     ('serdmanczyk@gmail.com', '2015-12-31T12:12:20', '0987901287340918723409', '18.00', '40.50', '50.2', '114.00', '50.0'),
--     ('serdmanczyk@gmail.com', '2015-12-31T12:12:22', '0987901287340918723409', '18.00', '40.20', '50.1', '114.10', '50.0'),
--     ('serdmanczyk@gmail.com', '2015-12-31T12:12:15', '9871623874612834768768', '18.00', '40.50', '50.2', '114.00', '50.0'),
--     ('serdmanczyk@gmail.com', '2015-12-31T12:12:16', '9871623874612834768768', '18.01', '40.51', '50.1', '114.00', '50.0'),
--     ('serdmanczyk@gmail.com', '2015-12-31T12:12:17', '9871623874612834768768', '18.00', '40.52', '50.1', '114.10', '50.0'),
--     ('serdmanczyk@gmail.com', '2015-12-31T12:12:18', '9871623874612834768768', '18.03', '40.53', '50.2', '114.00', '50.0'),
--     ('serdmanczyk@gmail.com', '2015-12-31T12:12:19', '9871623874612834768768', '18.00', '40.51', '50.0', '114.00', '50.0'),
--     ('serdmanczyk@gmail.com', '2015-12-31T12:12:20', '9871623874612834768768', '18.05', '40.50', '50.0', '114.00', '50.0'),
--     ('serdmanczyk@gmail.com', '2015-12-31T12:12:21', '9871623874612834768768', '18.00', '40.52', '50.2', '114.00', '50.0'),
--     ('sjohn@gmail.com', '2015-12-31T12:15:20', '6871623874612834768768', '18.05', '40.50', '50.0', '114.00', '50.0'),
--     ('sjohn@gmail.com', '2015-12-31T12:15:21', '6871623874612834768768', '18.05', '40.50', '50.0', '114.00', '50.0'),
--     ('sjohn@gmail.com', '2015-12-31T12:15:22', '6871623874612834768768', '18.05', '40.50', '50.0', '114.00', '50.0'),
--     ('sjohn@gmail.com', '2015-12-31T12:15:23', '6871623874612834768768', '18.05', '40.50', '50.0', '114.00', '50.0'),
--     ('sjohn@gmail.com', '2015-12-31T12:15:24', '6871623874612834768768', '18.05', '40.50', '50.0', '114.00', '50.0'),
--     ('sjohn@gmail.com', '2015-12-31T12:15:25', '6871623874612834768768', '18.05', '40.50', '50.0', '114.00', '50.0'),
--     ('sjohn@gmail.com', '2015-12-31T12:15:20', '4871623874612834768768', '18.05', '40.50', '50.0', '114.00', '50.0'),
--     ('sjohn@gmail.com', '2015-12-31T12:15:21', '4871623874612834768768', '18.05', '40.50', '50.0', '114.00', '50.0'),
--     ('sjohn@gmail.com', '2015-12-31T12:15:22', '4871623874612834768768', '18.05', '40.50', '50.0', '114.00', '50.0'),
--     ('sjohn@gmail.com', '2015-12-31T12:15:23', '4871623874612834768768', '18.05', '40.50', '50.0', '114.00', '50.0'),
--     ('sjohn@gmail.com', '2015-12-31T12:15:24', '4871623874612834768768', '18.05', '40.50', '50.0', '114.00', '50.0'),
--     ('sjohn@gmail.com', '2015-12-31T12:15:25', '4871623874612834768768', '18.05', '40.50', '50.0', '114.00', '50.0');

-- select readings.useremail, readings.posted, readings.coreid, readings.posted,
--     readings.temperature, readings.humidity, readings.moisture, readings.light, readings.battery
-- from readings inner join 
--     (select coreid, max(posted) from
--         (select * from readings where useremail = 'serdmanczyk@gmail.com') as userreadings
--     group by coreid)
-- as maxposted on readings.coreid = maxposted.coreid and readings.posted = maxposted.max;
