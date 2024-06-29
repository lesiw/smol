insert into baseurl(baseurl) values($1)
    on conflict (id) do update set baseurl = EXCLUDED.baseurl;
