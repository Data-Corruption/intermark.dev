#!/usr/bin/env node
/**
 * build-search.js
 *
 * Generate Lunr search index from pre-parsed documents.
 * Requirements: Node ≥14, lunr
 */

const fs   = require('fs')
const path = require('path')
const lunr = require('../../../assets/js/lunr.js')
const docsPath  = path.resolve(__dirname, '../../../public/.meta/search-pre-index.json')
const indexPath = path.resolve(__dirname, '../../../public/.meta/search-index.json')

// load the pre-generated documents
const docs = JSON.parse(fs.readFileSync(docsPath, 'utf8'))

// build
const idx = lunr(function () {
  this.ref('id')
  this.field('title', { boost: 10 })
  this.field('body')
  docs.forEach(d => this.add(d))
})

// write
fs.writeFileSync(
  indexPath,
  JSON.stringify({ index: idx }, null, 2),
  'utf8'
)