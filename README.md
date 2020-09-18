# WorkFusion
This activity allows you to copy and run a WorkFusion business process.

## Installation

### Flogo CLI
```bash
flogo install github.com/ecoletibco/workfusion
```

## Configuration

### Settings: 
| Name       | Type   | Description
|:---        | :---   | :---   
| url        | string | The URL used to connect to the WorkFusion API - ***REQUIRED***
| username   | string | The username used to connect to the WorkFusion API - ***REQUIRED***
| password   | string | The password used to connect to the WorkFusion API - ***REQUIRED***


### Input:
| Name  | Type   | Description
|:---   | :---   | :--- 
| uuid  | string | The UUID of the business process to copy and run 

### Output:
| Name  | Type   | Description
|:---   | :---   | :---     
| uuid  | string | The UUID of the new business process
| data  | any    | The final results data
