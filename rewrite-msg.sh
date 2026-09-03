#!/bin/sh
msg=$(cat)
case "$msg" in
  *"Initial commit: WhaleShop Go service"*)
    printf '%s\n' '初始提交：WhaleShop Go 服务'
    ;;
  *"docs: record WhaleShop deployment learnings"*)
    printf '%s\n' '文档：记录 WhaleShop 部署学习经验'
    ;;
  *"chore: move deployment notes to study log"*)
    printf '%s\n' '维护：将部署记录移至学习日志'
    ;;
  *)
    printf '%s' "$msg"
    ;;
esac
