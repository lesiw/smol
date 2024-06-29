insert into url(id, url)
    values ($1, $2)
    on conflict (id)
        do update set url = EXCLUDED.url;
