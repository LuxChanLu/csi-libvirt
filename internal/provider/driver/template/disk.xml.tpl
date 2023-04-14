<disk type="block" device="disk">
  <alias name="{{ .Alias }}"/>
  <driver name="qemu" type="raw"/>
  <source dev="{{ .Source }}"/>
  <target dev="{{ .Dev }}" bus="{{ .Bus }}"/>
  <serial>{{ .Serial }}</serial>
</disk>