# Varanus Requirements

Documentation of the software requirements for the Varanus monitor.

## Definitions

**system** - the Varanus monitor and all its components collectively
**target server** - a server whose health or continued function is monitored by the system
**violation** - a failure of a monitored system to perform as expected

## Requirements

### Configuration

1. The system shall accept a YAML configuration file defining the system configuration to be used.
2. The system configuration shall be validated when loaded by the system.
3. The system shall report validation errors to the user.
4. The system shall reject configurations that fail to pass validation.

### Security

1. Passwords and other sensitive configuration information shall be sealed using public key encryption.
2. The keys to unseal secrets in varanus shall be provided by a mechanism other than the configuration file.
3. Secrets shall be unsealed as close to time of use as possible.
4. Unsealed secrets shall be discarded as soon as possible after use.
5. The system shall provide a CLI utility for sealing configurations.

### Email Monitoring

1. The system shall monitor a target email server for end-to-end liveness by: 
   1. Sending an email from a separate account to a designated account on the target server
   2. Waiting a defined period for the email to arrive
   3. Checking that the email has arrived.
2. If the arrival check (#1.3) fails to arrive, the system shall retry the arrival check a designated number of times defined by the configuration.
3. If the arrival check fails all retrys, the system shall create a violation.

### Notification and Reporting

1. The system shall report violations using the contact methods defined in the configuration.
2. The system shall have the capability to report violations via email.
3. The system shall have the capability to report violations via SMS gateway.
