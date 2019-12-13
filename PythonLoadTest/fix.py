

""""Select 1 as player_id, 
        (select count(id) from player_statistics_entries where player_id = 1 and win = 'true') as wins, 
        (select count(id) from player_statistics_entries where player_id = 1 and  win = 'false') as loses,
        (select sum(points) from player_statistics_entries where player_id = 1 ) as points;
        """

        Select 1 as player_id, (select count(id) from player_statistics_entries where player_id = 1 and win = 'true') as wins, (select count(id) from player_statistics_entries where player_id = 1 and  win = 'false') as loses,(select sum(points) from player_statistics_entries where player_id = 1 ) as points;


        select player_id, sum(case when win = 'true' then 1 else 0 end) as wins, sum(case when win = 'false' then 1 else 0 end) as loses, sum(points) as points  from player_statistics_entries group by player_id order by points desc;




        Select 1 as player_id, (select count(id) from player_statistics_entries where player_id = 1 and win = 'true') as wins,(select count(id) from player_statistics_entries where player_id = 1 and  win = 'false') as loses,(select sum(points) from player_statistics_entries where player_id = 1 ) as points; 