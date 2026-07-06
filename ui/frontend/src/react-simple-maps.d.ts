declare module "react-simple-maps" {
  import { ComponentType, ReactNode } from "react";

  interface ComposableMapProps {
    projectionConfig?: { scale?: number; rotation?: number[] };
    style?: Record<string, string | number>;
    width?: number;
    height?: number;
    children?: ReactNode;
  }

  interface GeographiesProps {
    geography: string | object;
    children: (data: {
      geographies: GeographyType[];
    }) => ReactNode;
  }

  interface GeographyType {
    rsmKey: string;
    properties: Record<string, any>;
    geometry: any;
    type: string;
    id: string;
  }

  interface GeographyProps {
    geography: GeographyType;
    fill?: string;
    stroke?: string;
    strokeWidth?: number;
    style?: Record<string, any>;
  }

  interface MarkerProps {
    coordinates: [number, number];
    children?: ReactNode;
  }

  export const ComposableMap: ComponentType<ComposableMapProps>;
  export const Geographies: ComponentType<GeographiesProps>;
  export const Geography: ComponentType<GeographyProps>;
  export const Marker: ComponentType<MarkerProps>;
}
