# GoConductor
Like a conductor managing an orchestra of services. 

This program spawns a series of multiple microservices and manages them
by reading the data from a file that contains something like:
```yaml
---
microservice_name:
  - file_path: /path/to/file
  - min_spawn: 1 #(default: 1 and can't be lower than 1)
  - max_spawn: Inf # (default is INT.MAX)
  - spawn_rule: 6 
  - kill_rule: 2  
  - 
...
```
- **file_path** - The relative filepath to the runnable program
- **min_spawn** - The minimum processes that the conductor can create (default: 1 and can't be lower than 1)
- **max_spawn** - The maximum number of processes that the conductor can create (default is INT.MAX)
- **spawn_rule** - If the ration between the division of the available_elements_in_queue over spawned_microservices_nb is higher than the **spawn_rule** than the conductor will spawn a new service
- **kill_rule** - If the ration between the division of the available_elements_in_queue over spawned_microservices_nb is lower than the **kill_rule** than the conductor will send a message to the associated RabbitMQ queue to shut down a service

The communication between the conductor and other services is done by using RabbitMQ. For each microservice cluster the queue name that will be used to communicate is the `microservice_name`.

The Conductor will spawn both the initial and additional microservices automatically, but it will not automatically kill a microservice unless the Conductor is shut down.
In order to shut down a spawn the service will need to shut itself down when receiving from the queue a message like:
```json
{
  "kill": true
}
```