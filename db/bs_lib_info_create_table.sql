create table t_bs_lib_info
(
    id             bigint unsigned                         not null comment '主键' primary key auto_increment,
    name           varchar(64)   default ''                not null comment 'lib名称,唯一',
    latest_ver     varchar(256)  default ''                not null comment '最新版本',
    release_ver    varchar(256)  default ''                not null comment '历史发行版本',
    download_count int unsigned  default 0                 not null comment '下载次数',
    description    varchar(1024) default ''                not null comment '库描述',
    homepage       varchar(512)  default ''                not null comment '库主页链接',
    source_code    varchar(512)  default ''                not null comment '源码下载链接',
    dependencies   varchar(4096) default ''                not null comment '依赖的其他库信息',
    author         varchar(64)   default ''                not null comment '库作者姓名',
    contributors   varchar(2048) default ''                not null comment '贡献者信息',
    stars          int unsigned  default 0                 not null comment '星标数量',
    watching       int unsigned  default 0                 not null comment '关注数量',
    fork_count     int unsigned  default 0                 not null comment '拉取数量',
    create_time    timestamp     default current_timestamp not null comment '创建时间',
    update_time    timestamp     default current_timestamp not null on update current_timestamp comment '最新更新时间',
    constraint uniq_name
        unique (name)
) engine = InnoDB
  default charset = utf8
    comment '开源库信息'
