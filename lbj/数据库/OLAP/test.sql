-- 创建物化视图
CREATE MATERIALIZED VIEW pili_hera.test_bw_nerve_03 TO pili_hera.bw_nerve_aggre
AS SELECT
    en_name,
    real_server_id,
    record_time,
    isp,
    avgState(in_bw) AS in_bw, -- 满足func+'-State'规则
    avgState(out_bw) AS out_bw,
    avgState(max_bw) AS max_bw
FROM pili_hera.bw_nerve_new
GROUP BY en_name, real_server_id,record_time,isp;

CREATE MATERIALIZED VIEW pili_hera.test_wide_bw_03 to bw_nerve_wide_o1
AS SELECT
    a.en_name as en_name,
    a.real_server_id as real_server_id,
    a.record_time as record_time,
    a.in_bw as in_bw,
    a.out_bw as out_bw,
    a.max_bw as max_bw,
    isp_map.isp_id as isp_id,
    isp_map.isp_name as isp_name
FROM 
    (SELECT 
        en_name,
        real_server_id,
        record_time,
        isp,
        avgMerge(in_bw) AS in_bw, -- 满足func+'-State'规则
        avgMerge(out_bw) AS out_bw,
        avgMerge(max_bw) AS max_bw
    FROM
        pili_hera.bw_nerve_aggre 
    GROUP BY en_name, real_server_id,record_time,isp
    ) a
join isp_map
on a.isp = isp_map.isp_id;

CREATE MATERIALIZED VIEW pili_hera.test_wide_bw_04 to bw_nerve_wide_o2
AS SELECT
    a.en_name as en_name,
    a.real_server_id as real_server_id,
    a.record_time as record_time,
    a.isp_id as isp_id,
    a.isp_name as isp_name,
    a.in_bw as in_bw,
    a.out_bw as out_bw,
    a.max_bw as max_bw,
    b.hostname as hostname,
    b.real_idc_id as real_idc_id,
    b.country as country,
    b.region as region,
    b.province as province,
    b.pay_type as pay_type,
    b.isp as isp,
    b.idc_type as idc_type,
    b.idc_status as idc_status,
    b.total_bandwidth as total_bandwidth,
    b.guaranteed_bandwidth as guaranteed_bandwidth,
    b.archive_date as archive_date
FROM bw_nerve_wide_o1 a 
join res_idc_day_history b 
on b.node_id = a.real_server_id and b.archive_date = toDate(a.record_time);

CREATE MATERIALIZED VIEW pili_hera.test_wide_bw_05 to tt_wide
AS SELECT
    a.tto2_id as tto2_id,
    a.t_id as t_id,
    b.tto1_id as tto1_id
FROM tt_o2 a 
join tt_o1 b 
on b.t_id = a.t_id;

-- 创建物化视图对应的存储实体表
create table bw_nerve_aggre (
`en_name` String,
`real_server_id` String,
`isp` Int32,
`record_time` DateTime,
`in_bw` AggregateFunction(avg,Float64),
`out_bw` AggregateFunction(avg,Float64),
`max_bw` AggregateFunction(avg,Float64)
)
ENGINE=AggregatingMergeTree()
PARTITION BY toYYYYMM(record_time)
ORDER BY record_time
SETTINGS index_granularity = 8192;

create table bw_nerve_wide_o1 (
`en_name` String,
`real_server_id` String,
`isp_id` Int32,
`isp_name` String,
`record_time` DateTime,
`in_bw` Float64,
`out_bw` Float64,
`max_bw` Float64,
)
ENGINE=MergeTree()
PARTITION BY toYYYYMM(record_time)
ORDER BY record_time
SETTINGS index_granularity = 8192;

create table bw_nerve_wide_o2 (
`en_name` String,
`real_server_id` String,
`isp_id` Int32,
`isp_name` String,
`in_bw` Float64,
`out_bw` Float64,
`max_bw` Float64,
`hostname` String,
`real_idc_id` String,
`country` String,
`region` String,
`province` String,
`pay_type` String,
`isp` String,
`idc_type` String,
`idc_status` String,
`dtime` Int64,
`total_bandwidth` Float64,
`guaranteed_bandwidth` Float64,
`archive_date` Date,
`record_time` DateTime
)
ENGINE=MergeTree()
PARTITION BY toYYYYMM(record_time)
ORDER BY record_time
SETTINGS index_granularity = 8192;

create table tt_o1 (
`tto1_id` Int32,
`t_id` Int32,
)
ENGINE=MergeTree()
ORDER BY t_id
SETTINGS index_granularity = 8192;

create table tt_o2 (
`tto2_id` Int32,
`t_id` Int32,
)
ENGINE=MergeTree()
ORDER BY t_id
SETTINGS index_granularity = 8192;

create table tt_wide (
`tto2_id` Int32,
`t_id` Int32,
`tto1_id` Int32,
)
ENGINE=MergeTree()
ORDER BY t_id
SETTINGS index_granularity = 8192;

-- 查询sql

select w.real_idc_id as real_idc_id,
	   w.isp as isp,
       max(w.in_bw) as peak_in,
	   max(w.out_bw) as peak_out, 
	   max(w.max_bw) as peak_max,
	   quantileExact(0.95)(w.out_bw) as peak_95,
	   quantileExact(0.995)(w.out_bw) as peak_995,
	   quantileExact(0.999)(w.out_bw) as peak_999,
	   count(1) as total_rows
from (
	select 
		t.real_idc_id as real_idc_id,
		b.en_name as en_name,
		b.isp as isp,
		b.record_time as record_time,
		b.in_bw as in_bw,
		b.out_bw as out_bw,
		b.max_bw as max_bw
	from (
		${V3_SWITCH_NODE_BW}
	) b 
	join (
		select distinct
			real_idc_id,
			hostname,
			real_server_id 
		from (
			${IDC_ARCH_NODE}
		) s
		where real_server_id like '%-switch'
	) t on (b.real_server_id = t.real_server_id)

	UNION ALL

	select 
		t.real_idc_id as real_idc_id,
		concat(t.real_idc_id, '-vswitch') as en_name,
		b.isp as isp,
		b.record_time as record_time,
		sum(b.in_bw) as in_bw,
		sum(b.out_bw) as out_bw,
		sum(b.max_bw) as max_bw
	from (
		${V3_ALL_NODE_BW}
	) b 
	join (
		${NO_SWITCH_NODE}
	) t on (b.real_server_id = t.real_server_id)
	group by t.real_idc_id, b.isp, b.record_time
) w
group by w.real_idc_id,w.isp

select w.real_idc_id as real_idc_id,
	   w.isp as isp,
       max(w.in_bw) as peak_in,
	   max(w.out_bw) as peak_out, 
	   max(w.max_bw) as peak_max,
	   quantileExact(0.95)(w.out_bw) as peak_95,
	   quantileExact(0.995)(w.out_bw) as peak_995,
	   quantileExact(0.999)(w.out_bw) as peak_999,
	   count(1) as total_rows
from (
	select 
		t.real_idc_id as real_idc_id,
		b.en_name as en_name,
		b.isp as isp,
		b.record_time as record_time,
		b.in_bw as in_bw,
		b.out_bw as out_bw,
		b.max_bw as max_bw
	from (
		select 
            record_time,
            real_server_id,
            en_name,
            isp,
            avg(in_bw) as in_bw,
            avg(out_bw) as out_bw,
            avg(max_bw) as max_bw
        from pili_hera.bw_nerve_new 
        where record_time  >= '2022-09-01 00:00:00' and record_time <= '2022-09-30 23:59:59'
        group by en_name, isp, record_time, real_server_id 
	) b 
	join (
		select distinct
			real_idc_id,
			hostname,
			real_server_id 
		from (
			select 
                distinct
                real_idc_id,
                hostname,
                isp,
                node_id as real_server_id 
            from pili_hera.res_idc_day_history
            where dtime=0 and archive_date >= '2022-09-01' and archive_date <= '2022-09-30'
		) s
		where real_server_id like '%-switch'
	) t on (b.real_server_id = t.real_server_id)
) w
group by w.real_idc_id,w.isp

select 
    w.real_idc_id as real_idc_id,
    w.isp as isp,
    max(w.in_bw) as peak_in,
    max(w.out_bw) as peak_out, 
    max(w.max_bw) as peak_max,
    quantileExact(0.95)(w.out_bw) as peak_95,
    quantileExact(0.995)(w.out_bw) as peak_995,
    quantileExact(0.999)(w.out_bw) as peak_999,
    count(1) as total_rows
from bw_nerve_wide_o2 w
where
real_server_id like '%-switch' and record_time >='2022-09-01' and record_time<'2022-10-01'
group by real_idc_id, isp

-- 建表问题
1. 有无switch的逻辑,是否能够通过加字段完成