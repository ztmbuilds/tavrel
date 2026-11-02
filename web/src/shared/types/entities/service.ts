export interface ServiceNode {
  id: string;
}

export interface ServiceEdge {
  from: string;
  to: string;
}

export interface ServiceGraph {
  nodes: ServiceNode[];
  edges: ServiceEdge[];
}
