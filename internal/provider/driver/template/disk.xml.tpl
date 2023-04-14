<disk type="file" device="disk">
  <alias name="{{ .Alias }}"/>
  <driver name="qemu" type="raw"/>
  <source file="{{ .Source }}"/>
  <target bus="{{ .Bus }}"/>
</disk>