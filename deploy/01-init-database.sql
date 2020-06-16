CREATE TABLE feeds(
	id serial PRIMARY KEY,
	name VARCHAR (255) NOT NULL,
	uri VARCHAR (255) UNIQUE NOT NULL
);

CREATE TABLE userfeeds(
    user_id INTEGER NOT NULL,
    feed_id INTEGER NOT NULL,

    PRIMARY KEY (user_id, feed_id),

    CONSTRAINT userfeeds_feed_fk FOREIGN KEY (feed_id)
      REFERENCES feeds (id) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE NO ACTION
);