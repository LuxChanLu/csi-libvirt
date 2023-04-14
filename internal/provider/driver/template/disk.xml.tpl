<disk type="file" device="disk">
  <alias name="{{ .Alias }}"/>
  <driver name="qemu" type="raw"/>
  <source dev="{{ .Source }}"/>
  <backingStore/>
  <target dev="{{ .Dev }}" bus="{{ .Bus }}"/>
</disk>