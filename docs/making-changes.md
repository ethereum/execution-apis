# Contributors Guide

This guide will explain for new and experienced contributors alike how to
propose changes to Ethereum JSON-RPC API.

## Introduction

The Ethereum JSON-RPC API is the canonical interface between users and the
Ethereum network. Each execution layer client implements the API as defined by
the spec (or aspires to). 

As the main source of chain information, anything that is not
provided over via API will not be easily accessible to users. 

## Guiding Principles

When considering a change to the API, it's important to keep a few guiding
principles in mind.

### Backwards Compatibility

There is currently no accepted path to making backwards incompatible changes to
the API. This means that proposals which change syntax or semantics of existing
methods are unlikely to be accepted.

Even changes that would traditionally be accepted as backwards compatibible,
like adding an optional parameter, are not currently possible. This is due
the inability of clients to declare what version of a method they support.

### Unopionated Implementation


## Standardization

### Idea

### Proposal

### Aquiring Support

### Accepting the Change

###
