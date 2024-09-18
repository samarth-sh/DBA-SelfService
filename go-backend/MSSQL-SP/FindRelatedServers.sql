CREATE OR ALTER PROCEDURE dbo.FindRelatedServers
    @ServerIP NVARCHAR(255),
    @RelatedServers NVARCHAR(MAX) OUTPUT
AS 
BEGIN

with cte as (
	select * 
	from dbo.sma_hadr_ag ag
	where 1=1
	and (ag.server = @ServerIP or ag.ag_listener_ip1 = @ServerIP or ag.ag_listener_ip2 = @ServerIP)
)
,cte2 as (
	select r.server
	from cte c
	left join dbo.sma_hadr_ag r
		on r.server <> c.server
		and ( r.ag_listener_ip1 = c.ag_listener_ip1 and  r.ag_listener_ip2 = c.ag_listener_ip2)
	union
	select c.server
	from cte c
)
select @RelatedServers = STRING_AGG(server, ',')
from cte2
END;
