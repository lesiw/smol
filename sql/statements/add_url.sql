with cte as (
    insert into url(id, url)
    select $1, $2 where not exists (select 1 from url where url = $2)
    on conflict (id) do nothing
    returning id
)
select id from url where url = $2
union all
select id from cte;
