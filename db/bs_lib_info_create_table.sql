create table t_bs_lib_info
(
    id             bigint unsigned auto_increment comment '主键'
        primary key,
    name           varchar(256)  default ''                not null comment 'lib名称',
    version        varchar(256)  default ''                not null comment '版本号',
    license        varchar(256)  default ''                not null comment '许可证',
    download_count int unsigned  default 0                 not null comment '下载次数',
    description    varchar(4096) default ''                not null comment '库描述',
    homepage       varchar(1024) default ''                not null comment '库主页链接',
    source_code    varchar(1024) default ''                not null comment '源码下载链接',
    dependencies   text                                    not null comment '依赖的其他库信息',
    author         varchar(1024) default ''                not null comment '库作者姓名',
    contributors   varchar(4096) default ''                not null comment '贡献者信息',
    stars          int unsigned  default 0                 not null comment '星标数量',
    watching       int unsigned  default 0                 not null comment '关注数量',
    fork_count     int unsigned  default 0                 not null comment '拉取数量',
    create_time    timestamp     default CURRENT_TIMESTAMP not null comment '创建时间',
    update_time    timestamp     default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '最新更新时间'
) engine = InnoDB default char set ='utf8mb4'
    comment '开源库信息';

create index t_bs_lib_info_name_version_idx
    on t_bs_lib_info (name, version);

