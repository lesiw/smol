package stmt

const AddUrl = `
with cte as (
    insert into url(id, url)
    select $1, $2 where not exists (select 1 from url where url = $2)
    on conflict (id) do nothing
    returning id
)
select id from url where url = $2
union all
select id from cte;
`

const GetDomain = `
select baseurl from baseurl limit 1;
`

const GetUrl = `
select url from url where id = $1
`

const SetDomain = `
insert into baseurl(baseurl) values($1)
    on conflict (id) do update set baseurl = EXCLUDED.baseurl;
`

const SetUrl = `
insert into url(id, url)
    values ($1, $2)
    on conflict (id)
        do update set url = EXCLUDED.url;
`

