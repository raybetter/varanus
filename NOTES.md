Varanus feature list:
- read configuration from a YAML or JSON file
- schedule monitor tasks at various intervals
- execute scheduled monitor tasks

Architecture

- generic monitoring event structure
  - can run -> returns true (do the event) or false (requeue the event)
  - test event -> execute the setup or test 



Configuration details

  - notifier modules
    - sms - configuration for SMS gateway
    - email - smtp credentials for sending notifications
  - notification groups
    - group name
    - list of notifications
      - notification type (SMS, email)
      - notification config (i.e destination number or email)
  - module configs
    - smtp
      - absolute maximum sending rate for a group of smtp servers
        - maybe the grouping is automatic by resolving the server IPs?
    - sms
      - absolute maximum sending rate for sms messages
    - imap
      - maximum polling frequency
  - report generation
    - report generation frequency
    - notifier group
  - monitor types
    - email e2e -> send a message from an smpt server and verify that it shows up at an imap server
      - config:
        - target email address
        - smpt sender config
          - server address
          - server port
          - server ssl
          - username & password
        - imap receiver config
          - server address
          - server port
          - server ssl
          - username & password
        - sending rate
    - dns monitor - check that DNS entry exists

TODOs:
- make an executable that can seal and unseal YAML config files
- varanus seal -config <yaml file> -key <public key> [-output <filename>]
  - output is optional, if omitted, replace the original file
  - if output is specified, won't overwrite an existing file
- varanus unseal -config <yaml file> -key <private key> -passphrase <passphrase value>] [-output <filename>]
  - passphrase is optional, support pass:string, file:filename, and env:var like openssl
  - output is optional, if omitted, replace the original file
  - if output is specified, won't overwrite an existing file

- ...

Design decisions:

- Configuration management
  - decided to use regular YAML parsing for config loading
    - viper, the main config library, doesn't have good support for deeply nested defaults, which
      defeats a lot of its purpose.  The main advantage would have been to parcel out config loading
      of each module to itself, but we can get most of this from the validation logic which we have
      to do anyway
    - the `gopkg.in/yaml.v3` library provides YAML parsing for the config
  - decided against using `github.com/creasty/defaults` setting defaults (e.g. port defaults) -- requiring explicit config for everything is probably better anyway.

Sealed secrets
  - memguard
    - looked at [memguard](https://github.com/awnumar/memguard), which provides sealed enclaves and
      locked buffers (locked memory pages)
    - The problem is that once you parse the locked buffer to an object, like rsa.PrivateKey, then
      the parse object is not protected.
    - For now, decide to trust the process with the private key loaded from the file

Walker
  - implement a recursive, depth-first walk down an object tree (e.g. walking items in maps and 
    lists and fields of structs).  
  - the walk can be implmented as mutable (expect to modify the needle objects in the callback) or
    immutable (read only).
  - the walk call accepts a `reflect.Type` type definition that can be an interface or a concrete
    type.  For a 
  - the call to the walk provides a callback function that has a `needle interface{}` argument that
    should be cast to a `*NeedleType` (for the mutable walk) or a `NeedleType`
  - tried creating a generic function that would do the type casting, but couldn't find a way to
    pass an interface argument -- the generic type inference fails.  So just keep them as
    `interface{}` types.
  - use the walker to streamline the implementation of validation and sealing in the config 
    management.
    
References:

- crypto
  - https://www.developer.com/languages/cryptography-in-go/
