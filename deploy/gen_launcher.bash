OUTPUT=$1

if [ "$OUTPUT" == "" ]
then
  echo $0 [output path]
  exit -1
fi

for f in ../bin/*
do
  NAME=${f##.*/}
  cat launcher.templ | sed "s/{{bin_name}}/$NAME/g" > "$OUTPUT/exfebus_$NAME"
  chmod +x "$OUTPUT/exfebus_$NAME"
done