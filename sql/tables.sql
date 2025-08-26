PRAGMA foreign_keys = ON;
-- users
create table if not exists users (
    uuid text not null primary key unique,
    username text not null,
    email text not null,
    password text not null,
    notregistered boolean not null,
    lastseen text not null,
    loggedin boolean not null
);

-- posts
create table if not exists posts (
    id integer primary key autoincrement,
    title text not null,
    content text not null,
    author_uuid text not null,
    foreign key(author_uuid) references users(uuid)
);

-- comments
create table if not exists comments (
    id integer primary key autoincrement,
    content text not null,
    comment_author_uuid text not null,
    post_id integer not null,
    foreign key(comment_author_uuid) references users(uuid),
    foreign key(post_id) references posts(id)
);

-- interactions 
create table if not exists interactions (
    id integer primary key autoincrement,
    user_uuid text not null,
    post_id integer not null,
    liked boolean not null default 0,
    disliked boolean not null default 0,
    foreign key(user_uuid) references users(uuid),
    foreign key(post_id) references posts(id)
);


