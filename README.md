
### Ejercicio N°1 Consigna:
Definir un script de bash `generar-compose.sh` que permita crear una definición de Docker Compose con una cantidad configurable de clientes.  El nombre de los containers deberá seguir el formato propuesto: client1, client2, client3, etc. 

El script deberá ubicarse en la raíz del proyecto y recibirá por parámetro el nombre del archivo de salida y la cantidad de clientes esperados:

`./generar-compose.sh docker-compose-dev.yaml 5`

Considerar que en el contenido del script pueden invocar un subscript de Go o Python:

```
#!/bin/bash
echo "Nombre del archivo de salida: $1"
echo "Cantidad de clientes: $2"
python3 mi-generador.py $1 $2
```

En el archivo de Docker Compose de salida se pueden definir volúmenes, variables de entorno y redes con libertad, pero recordar actualizar este script cuando se modifiquen tales definiciones en los sucesivos ejercicios.

### Ejercicio N°1 explicación de solución:
Se creó un script muy sencillo de Bash, que solo se encarga de imprimir los parámetros por consola y luego enviarselos a un archivo de python. Este archivo llamado "mi-generador.py" es el encargado de:
1-Validar que son parámetros válidos para generar un nuevo archivo de compose.
2-Abrir un archivo nuevo, con el nombre especificado.
3-Agregar al servidor al archivo, siguiendo el formato del servidor de ejemplo.
4-Agregar tantos clientes como se le especificaron, siguiendo el formato anterior y asegurándose de cambiar sus ids y nombres.
5-Agregar la red a la cual están todos conectados.
6-Si pudo agregar todo y guardar el archivo con éxito, notifica al usuario y el programa termina.

### Dificultades Encontradas
-Hubo ciertas complicaciones a la hora de parsear el archivo, a veces las líneas quedaban desfasadas.