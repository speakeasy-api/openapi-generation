# Vehicle

A non-discriminated union of vehicle types in the bar namespace


## Supported Types

### Car

```go
vehicle := examplealias.NewVehicle(bar.Car{/* values here */})
```

### Bike

```go
vehicle := examplealias.NewVehicle(bar.Bike{/* values here */})
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch vehicle.Type {
	case examplealias.VehicleTypeCar:
		// vehicle.Car is populated
	case examplealias.VehicleTypeBike:
		// vehicle.Bike is populated
	default:
		// Unknown type - use vehicle.GetUnknownRaw() for raw JSON
}
```
