import cytoscape, { LayoutOptions, NodeSingular } from 'cytoscape'
import cola from 'cytoscape-cola'
import dagre from 'cytoscape-dagre'
import fcose from 'cytoscape-fcose'
import { FcoseLayoutOptions } from 'cytoscape-fcose'

import { useEffect, useRef, useState } from 'react'
import type React from 'react'
import { useStreamModules } from '../../api/modules/use-stream-modules'
import { useUserPreferences } from '../../providers/user-preferences-provider'
import { createGraphStyles } from './graph-styles'
import { type FTLNode, getGraphData } from './graph-utils'

cytoscape.use(cola)
cytoscape.use(fcose)
cytoscape.use(dagre)

interface GraphPaneProps {
  onTapped?: (item: FTLNode | null, moduleName: string | null) => void
}

const ZOOM_THRESHOLD = 0.6

export const GraphPane: React.FC<GraphPaneProps> = ({ onTapped }) => {
  const modules = useStreamModules()
  const { isDarkMode } = useUserPreferences()

  const cyRef = useRef<HTMLDivElement>(null)
  const cyInstance = useRef<cytoscape.Core | null>(null)
  const [nodePositions, setNodePositions] = useState<Record<string, { x: number; y: number }>>({})
  const resizeObserverRef = useRef<ResizeObserver | null>(null)

  // Initialize Cytoscape and ResizeObserver
  useEffect(() => {
    if (!cyRef.current) return

    cyInstance.current = cytoscape({
      container: cyRef.current,
      userZoomingEnabled: true,
      userPanningEnabled: true,
      boxSelectionEnabled: false,
      autoungrabify: true,
    })

    // Create ResizeObserver
    resizeObserverRef.current = new ResizeObserver(() => {
      if (!cyInstance.current) return

      cyInstance.current?.resize()
    })

    // Start observing the container
    resizeObserverRef.current.observe(cyRef.current)

    // Add click handlers
    cyInstance.current.on('tap', 'node', (evt) => {
      const node = evt.target
      const nodeType = node.data('type')
      const item = node.data('item')
      const zoom = evt.cy.zoom()
      const moduleName = node.parent().length ? node.parent().data('id') : node.data('id')

      if (zoom < ZOOM_THRESHOLD) {
        if (nodeType === 'node') {
          const parent = node.parent()
          if (parent.length) {
            onTapped?.(parent.data('item'), parent.data('id'))
            return
          }
        }
      }

      if (nodeType === 'groupNode' || (nodeType === 'node' && zoom >= ZOOM_THRESHOLD)) {
        onTapped?.(item, moduleName)
      }
    })

    cyInstance.current.on('tap', (evt) => {
      if (evt.target === cyInstance.current) {
        onTapped?.(null, null)
      }
    })

    // Update zoom level event handler
    cyInstance.current.on('zoom', (evt) => {
      const zoom = evt.target.zoom()
      const elements = evt.target.elements()

      if (zoom < ZOOM_THRESHOLD) {
        // Hide child nodes
        elements.nodes('node[type != "groupNode"]').style('opacity', 0)

        // Show only module-level edges (type="moduleConnection")
        elements.edges('[type = "moduleConnection"]').style('opacity', 1)
        elements.edges('[type = "childConnection"]').style('opacity', 0)

        // Updated text settings for zoomed out view
        elements.nodes('node[type = "groupNode"]').style({
          'text-valign': 'center',
          'text-halign': 'center',
          'font-size': '18px',
          'text-max-width': '160px',
          'text-margin-y': '0px',
          width: '180px',
        })
      } else {
        // Show all nodes
        elements.nodes().style('opacity', 1)

        // Show only verb-level edges (type="childConnection")
        elements.edges('[type = "moduleConnection"]').style('opacity', 0)
        elements.edges('[type = "childConnection"]').style('opacity', 1)

        // Move text to top when zoomed in
        elements.nodes('node[type = "groupNode"]').style({
          'text-valign': 'top',
          'text-halign': 'center',
          'font-size': '14px',
          'text-max-width': '160px',
          'text-margin-y': '-10px',
          width: '180px',
        })
      }
    })

    return () => {
      if (resizeObserverRef.current) {
        resizeObserverRef.current.disconnect()
      }
      cyInstance.current?.destroy()
    }
  }, [])

  // Modify the data loading effect
  useEffect(() => {
    if (!cyInstance.current) return

    const elements = getGraphData(modules.data, isDarkMode, nodePositions)
    const cy = cyInstance.current

    // Update existing elements and add new ones
    const nodes = elements.filter((element) => element.group === 'nodes')
    const edges = elements.filter((element) => element.group === 'edges')

    // First handle nodes
    for (const element of nodes) {
      const id = element.data?.id
      if (!id) continue // Skip elements without an id

      const existingElement = cy.getElementById(id)

      if (existingElement.length) {
        // Update existing element data
        existingElement.data(element.data)

        // If it's a node and doesn't have saved position, update position
        if (!nodePositions[id]) {
          existingElement.position(element.position || { x: 0, y: 0 })
        }
      } else {
        // Add new element
        cy.add(element)
      }
    }

    // Then handle edges after all nodes exist
    for (const element of edges) {
      const id = element.data?.id
      if (!id) continue

      const existingElement = cy.getElementById(id)

      if (existingElement.length) {
        existingElement.data(element.data)
      } else {
        cy.add(element)
      }
    }

    // Remove elements that no longer exist in the data
    for (const element of cy.elements()) {
      const elementId = element.data('id')
      const stillExists = elements.some((e) => e.data?.id === elementId)
      if (!stillExists) {
        element.remove()
      }
    }

    // Only run layout for new nodes without positions
    const hasNewNodesWithoutPositions = cy.nodes().some((node) => {
      const nodeId = node.data('id')
      return node.data('type') === 'groupNode' && !nodePositions[nodeId]
    })

    if (hasNewNodesWithoutPositions) {
      /**
       * TODO The multiple layout options here are for demo purposes. A single layout should be selected and the unused ones removed.
       */
      // @ts-ignore
      var colaLayoutOptions = {
        name: 'cola',
        animate: true, // whether to show the layout as it's running
        refresh: 1, // number of ticks per frame; higher is faster but more jerky
        maxSimulationTime: 4000, // max length in ms to run the layout
        ungrabifyWhileSimulating: false, // so you can't drag nodes during layout
        fit: true, // on every layout reposition of nodes, fit the viewport
        padding: 30, // padding around the simulation
        boundingBox: undefined, // constrain layout bounds; { x1, y1, x2, y2 } or { x1, y1, w, h }
        nodeDimensionsIncludeLabels: false, // whether labels should be included in determining the space used by a node

        // layout event callbacks
        ready: function () {}, // on layoutready
        stop: function () {}, // on layoutstop

        // positioning options
        randomize: false, // use random node positions at beginning of layout
        avoidOverlap: true, // if true, prevents overlap of node bounding boxes
        handleDisconnected: true, // if true, avoids disconnected components from overlapping
        convergenceThreshold: 0.01, // when the alpha value (system energy) falls below this value, the layout stops
        nodeSpacing: function () {
          return 10
        }, // extra spacing around nodes
        flow: undefined, // use DAG/tree flow layout if specified, e.g. { axis: 'y', minSeparation: 30 }
        alignment: undefined, // relative alignment constraints on nodes, e.g. {vertical: [[{node: node1, offset: 0}, {node: node2, offset: 5}]], horizontal: [[{node: node3}, {node: node4}], [{node: node5}, {node: node6}]]}
        gapInequalities: undefined, // list of inequality constraints for the gap between the nodes, e.g. [{"axis":"y", "left":node1, "right":node2, "gap":25}]
        centerGraph: true, // adjusts the node positions initially to center the graph (pass false if you want to start the layout from the current position)

        // different methods of specifying edge length
        // each can be a constant numerical value or a function like `function( edge ){ return 2; }`
        edgeLength: undefined, // sets edge length directly in simulation
        edgeSymDiffLength: undefined, // symmetric diff edge length in simulation
        edgeJaccardLength: undefined, // jaccard edge length in simulation

        // iterations of cola algorithm; uses default values on undefined
        unconstrIter: undefined, // unconstrained initial layout iterations
        userConstIter: undefined, // initial layout iterations with user-specified constraints
        allConstIter: undefined, // initial layout iterations with all constraints including non-overlap
      } as LayoutOptions

      // @ts-ignore
      var dagreLayoutOptions = {
        name: 'dagre',
        // dagre algo options, uses default value on undefined
        nodeSep: undefined, // the separation between adjacent nodes in the same rank
        edgeSep: undefined, // the separation between adjacent edges in the same rank
        rankSep: undefined, // the separation between each rank in the layout
        rankDir: undefined, // 'TB' for top to bottom flow, 'LR' for left to right,
        align: undefined, // alignment for rank nodes. Can be 'UL', 'UR', 'DL', or 'DR', where U = up, D = down, L = left, and R = right
        acyclicer: undefined, // If set to 'greedy', uses a greedy heuristic for finding a feedback arc set for a graph.
        // A feedback arc set is a set of edges that can be removed to make a graph acyclic.
        ranker: undefined, // Type of algorithm to assign a rank to each node in the input graph. Possible values: 'network-simplex', 'tight-tree' or 'longest-path'
        minLen: function () {
          return 1
        }, // number of ranks to keep between the source and target of the edge
        edgeWeight: function () {
          return 1
        }, // higher weight edges are generally made shorter and straighter than lower weight edges

        // general layout options
        fit: true, // whether to fit to viewport
        padding: 60, // fit padding
        spacingFactor: 1.5, // Applies a multiplicative factor (>0) to expand or compress the overall area that the nodes take up
        nodeDimensionsIncludeLabels: false, // whether labels should be included in determining the space used by a node
        animate: true, // whether to transition the node positions
        animateFilter: function () {
          return true
        }, // whether to animate specific nodes when animation is on; non-animated nodes immediately go to their final positions
        animationDuration: 500, // duration of animation in ms if enabled
        animationEasing: undefined, // easing of animation if enabled
        boundingBox: undefined, // constrain layout bounds; { x1, y1, x2, y2 } or { x1, y1, w, h }
        transform: function (_node: NodeSingular, pos: cytoscape.Position) {
          return pos
        }, // a function that applies a transform to the final node position
        ready: function () {}, // on layoutready
        sort: undefined, // a sorting function to order the nodes and edges; e.g. function(a, b){ return a.data('weight') - b.data('weight') }
        // because cytoscape dagre creates a directed graph, and directed graphs use the node order as a tie breaker when
        // defining the topology of a graph, this sort function can help ensure the correct order of the nodes/edges.
        // this feature is most useful when adding and removing the same nodes and edges multiple times in a graph.
        stop: function () {}, // on layoutstop
      } as LayoutOptions

      // @ts-ignore
      const fcoseLayoutOptions = {
        name: 'fcose',
        animate: false,
        quality: 'proof',
        nodeSeparation: 150,
        idealEdgeLength: 200,
        nodeRepulsion: 20000,
        padding: 50,
        randomize: false,
        tile: true,
        tilingPaddingVertical: 100,
        tilingPaddingHorizontal: 100,
        fit: true,
        componentSpacing: 100,
        edgeElasticity: 0.45,
        gravity: 0.75,
        initialEnergyOnIncremental: 0.5,
      } as FcoseLayoutOptions

      // const layout = cy.layout(dagreLayoutOptions)
      // const layout = cy.layout(fcoseLayoutOptions)
      const layout = cy.layout(colaLayoutOptions)
      layout.run()

      layout.on('layoutstop', () => {
        const newPositions = { ...nodePositions }
        for (const node of cy.nodes()) {
          const nodeId = node.data('id')
          newPositions[nodeId] = node.position()
        }
        setNodePositions(newPositions)
      })
    }
  }, [nodePositions, modules.data, isDarkMode])

  useEffect(() => {
    // Update your cytoscape instance with new styles when dark mode changes
    cyInstance.current?.style(createGraphStyles(isDarkMode))
  }, [isDarkMode])

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative', minWidth: 0 }}>
      <div
        ref={cyRef}
        style={{
          width: '100%',
          height: '100%',
          position: 'absolute',
          zIndex: 0, // Ensure graph stays below other elements
        }}
      />
    </div>
  )
}
