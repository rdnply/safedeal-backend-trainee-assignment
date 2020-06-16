CREATE TABLE products (
	id SERIAL PRIMARY KEY,
	name VARCHAR (150) NOT NULL,
	width DOUBLE PRECISION NOT NULL,
	length DOUBLE PRECISION NOT NULL,
	height DOUBLE PRECISION NOT NULL,
	weight DOUBLE PRECISION NOT NULL,
	place VARCHAR (200) NOT NULL
)

CREATE TABLE orders (
	id SERIAL PRIMARY KEY,
	product_id INTEGER REFERENCES products (id) NOT NULL,
	name VARCHAR (150) NOT NULL,
	from_place VARCHAR (200) NOT NULL,
	destination VARCHAR (200) NOT NULL,
	time TIMESTAMP WITH TIME ZONE NOT NULL
)